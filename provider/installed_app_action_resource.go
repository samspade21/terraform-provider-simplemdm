package provider

import (
	"context"
	"fmt"
	"time"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &installedAppActionResource{}
	_ resource.ResourceWithConfigure = &installedAppActionResource{}
)

type installedAppActionResource struct {
	client *simplemdm.Client
}

type installedAppActionModel struct {
	ID             types.String `tfsdk:"id"`
	InstalledAppID types.String `tfsdk:"installed_app_id"`
	Action         types.String `tfsdk:"action"`
	LastTriggered  types.String `tfsdk:"last_triggered"`
}

func InstalledAppActionResource() resource.Resource {
	return &installedAppActionResource{}
}

func (r *installedAppActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_installed_app_action"
}

func (r *installedAppActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a one-shot action against an installed app on a device. Supported actions: 'update' (push update), 'request_management' (request MDM management of an unmanaged app), 'uninstall' (DELETE the install record).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Synthetic identifier `<installed_app_id>:<action>:<unix_nanos>`.",
			},
			"installed_app_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required. Installed app ID.",
			},
			"action": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("update", "request_management", "uninstall"),
				},
				Description: "Required. One of `update`, `request_management`, `uninstall`.",
			},
			"last_triggered": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *installedAppActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

func (r *installedAppActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan installedAppActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.InstalledAppID.ValueString()
	var err error
	switch plan.Action.ValueString() {
	case "update":
		err = simplemdmext.PushInstalledAppUpdate(ctx, r.client, id)
	case "request_management":
		err = simplemdmext.RequestInstalledAppManagement(ctx, r.client, id)
	case "uninstall":
		err = simplemdmext.DeleteInstalledApp(ctx, r.client, id)
	default:
		resp.Diagnostics.AddError("Invalid action", fmt.Sprintf("unknown action %q", plan.Action.ValueString()))
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Failed to invoke installed app action", err.Error())
		return
	}

	now := time.Now()
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%d", id, plan.Action.ValueString(), now.UnixNano()))
	plan.LastTriggered = types.StringValue(now.UTC().Format(time.RFC3339))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *installedAppActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state installedAppActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *installedAppActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan installedAppActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *installedAppActionResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
