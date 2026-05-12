package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// assignmentGroupAppBindingResource lets downstream Terraform attach an
// existing app to an existing assignment group without taking ownership of
// either resource. This is the binding-only counterpart to the apps map on
// simplemdm_assignmentgroup; it exists so consumers can manage memberships of
// UI-owned assignment groups (notably dynamic ones whose membership rules
// aren't API-exposed).
type assignmentGroupAppBindingResource struct {
	client *simplemdm.Client
}

type assignmentGroupAppBindingModel struct {
	ID                types.String `tfsdk:"id"`
	AssignmentGroupID types.String `tfsdk:"assignment_group_id"`
	AppID             types.String `tfsdk:"app_id"`
	DeploymentType    types.String `tfsdk:"deployment_type"`
	InstallType       types.String `tfsdk:"install_type"`
}

var (
	_ resource.Resource                = &assignmentGroupAppBindingResource{}
	_ resource.ResourceWithConfigure   = &assignmentGroupAppBindingResource{}
	_ resource.ResourceWithImportState = &assignmentGroupAppBindingResource{}
)

func AssignmentGroupAppBindingResource() resource.Resource {
	return &assignmentGroupAppBindingResource{}
}

func (r *assignmentGroupAppBindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assignmentgroup_app_binding"
}

func (r *assignmentGroupAppBindingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Binds an existing SimpleMDM app to an existing assignment group with optional deployment_type and install_type overrides. " +
			"Useful for managing memberships of UI-owned assignment groups (including dynamic groups) without taking ownership of the assignment group itself.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<assignment_group_id>:<app_id>\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"assignment_group_id": schema.StringAttribute{
				Required:    true,
				Description: "Identifier of the assignment group that should receive the app.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_id": schema.StringAttribute{
				Required:    true,
				Description: "Identifier of the app to bind to the assignment group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deployment_type": schema.StringAttribute{
				Optional: true,
				Description: "Deployment type override sent on assignment. Must be one of \"standard\" or \"munki\". " +
					"Mirrors the `apps_deployment_types` map on simplemdm_assignmentgroup. Changing the value " +
					"re-issues the assign call against SimpleMDM (no resource replacement needed).",
				Validators: []validator.String{
					stringvalidator.OneOf("standard", "munki"),
				},
			},
			"install_type": schema.StringAttribute{
				Optional: true,
				Description: "Install type override sent on assignment. Must be one of \"managed\", \"self_serve\", \"default_installs\", or \"managed_updates\". " +
					"Only meaningful when deployment_type == \"munki\". Mirrors the `apps_install_types` map on simplemdm_assignmentgroup.",
				Validators: []validator.String{
					stringvalidator.OneOf("managed", "self_serve", "default_installs", "managed_updates"),
				},
			},
		},
	}
}

func (r *assignmentGroupAppBindingResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*simplemdm.Client)
}

func (r *assignmentGroupAppBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan assignmentGroupAppBindingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.assignApp(plan); err != nil {
		resp.Diagnostics.AddError("Error binding app to assignment group", err.Error())
		return
	}

	plan.ID = types.StringValue(buildAssignmentGroupBindingID(plan.AssignmentGroupID.ValueString(), plan.AppID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *assignmentGroupAppBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state assignmentGroupAppBindingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Strategy: SimpleMDM's /api/v1/apps/{id} response carries no
	// relationships block (verified via probe), so we cannot do the
	// app-side inverse lookup we use for profiles. Instead, fetch the
	// assignment group itself and check `relationships.apps.data` for our
	// app ID. /api/v1/assignment_groups/{id} 404s when the group is gone,
	// which we treat as drift (remove the binding from state).
	//
	// We deliberately don't try to round-trip deployment_type / install_type
	// from the API: SimpleMDM doesn't expose per-app overrides on either GET
	// endpoint, so any read-back would always show "unknown" and force
	// repeated diffs. The state we wrote in Create/Update is the source of
	// truth for those fields.
	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s", r.client.HostName, state.AssignmentGroupID.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error building SimpleMDM assignment group request", err.Error())
		return
	}

	body, err := r.client.RequestResponse200(httpReq)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SimpleMDM assignment group", err.Error())
		return
	}

	bound, err := assignmentGroupHasAppBinding(body, state.AppID.ValueString(), state.AssignmentGroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error parsing SimpleMDM assignment group relationships", err.Error())
		return
	}
	if !bound {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(buildAssignmentGroupBindingID(state.AssignmentGroupID.ValueString(), state.AppID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *assignmentGroupAppBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state assignmentGroupAppBindingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only deployment_type and install_type are mutable (the IDs are
	// ForceNew). SimpleMDM ignores deployment_type / install_type on a POST
	// to an already-assigned app, so emulate the assignment-group resource's
	// proven pattern: unassign, then re-assign with the new params. The
	// transient gap is acceptable for the typical UI-owned dynamic-AG use
	// case where the override is rarely changed in steady state.
	if plan.DeploymentType.ValueString() != state.DeploymentType.ValueString() ||
		plan.InstallType.ValueString() != state.InstallType.ValueString() {
		if err := r.client.AssignmentGroupUnAssignApp(plan.AssignmentGroupID.ValueString(), plan.AppID.ValueString()); err != nil && !isNotFoundError(err) {
			resp.Diagnostics.AddError("Error un-assigning app prior to override update", err.Error())
			return
		}
		if err := r.assignApp(plan); err != nil {
			resp.Diagnostics.AddError("Error re-assigning app with new overrides", err.Error())
			return
		}
	}

	plan.ID = types.StringValue(buildAssignmentGroupBindingID(plan.AssignmentGroupID.ValueString(), plan.AppID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *assignmentGroupAppBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state assignmentGroupAppBindingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AssignmentGroupUnAssignApp(state.AssignmentGroupID.ValueString(), state.AppID.ValueString()); err != nil {
		if isNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Error removing app binding", err.Error())
	}
}

func (r *assignmentGroupAppBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	agID, appID, err := parseAssignmentGroupBindingID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected import identifier format",
			"Expected assignment_group_id:app_id or assignment_group_id|app_id, got: "+req.ID,
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("assignment_group_id"), agID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), appID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), buildAssignmentGroupBindingID(agID, appID))...)
}

// assignApp issues the SimpleMDM "Assign App" call with the plan's optional
// deployment_type and install_type query params. Reuses the shared client
// helper so behaviour matches simplemdm_assignmentgroup's apps_deployment_types
// / apps_install_types handling.
func (r *assignmentGroupAppBindingResource) assignApp(plan assignmentGroupAppBindingModel) error {
	return r.client.AssignmentGroupAssignApp(
		plan.AssignmentGroupID.ValueString(),
		plan.AppID.ValueString(),
		plan.DeploymentType.ValueString(),
		plan.InstallType.ValueString(),
	)
}

// assignmentGroupHasAppBinding inspects a /api/v1/assignment_groups/{id}
// response body and reports whether `relationships.apps.data` contains the
// given app ID. IDs come back as JSON numbers; decoding into json.Number
// rather than the default float64 avoids scientific-notation truncation for
// IDs > 1e6, which would otherwise silently miss real bindings.
func assignmentGroupHasAppBinding(body []byte, appID, assignmentGroupID string) (bool, error) {
	type relRef struct {
		ID json.Number `json:"id"`
	}
	type payload struct {
		Data struct {
			Relationships struct {
				Apps struct {
					Data []relRef `json:"data"`
				} `json:"apps"`
			} `json:"relationships"`
		} `json:"data"`
	}

	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		return false, fmt.Errorf("error parsing assignment group payload for group %s: %w", assignmentGroupID, err)
	}

	for _, a := range p.Data.Relationships.Apps.Data {
		if a.ID.String() == appID {
			return true, nil
		}
	}
	return false, nil
}
