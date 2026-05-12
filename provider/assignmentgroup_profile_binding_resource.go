package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// assignmentGroupProfileBindingResource lets downstream Terraform attach an
// existing profile (or custom declaration) to an existing assignment group
// without taking ownership of either resource. This is the binding-only
// counterpart to simplemdm_assignmentgroup; the parent AG itself can be
// UI-managed (including dynamic groups whose membership rules are not
// API-exposed) and any profile type that lives at /api/v1/profiles will work
// because the same join endpoint accepts profile and declaration IDs alike.
type assignmentGroupProfileBindingResource struct {
	client *simplemdm.Client
}

type assignmentGroupProfileBindingModel struct {
	ID                types.String `tfsdk:"id"`
	AssignmentGroupID types.String `tfsdk:"assignment_group_id"`
	ProfileID         types.String `tfsdk:"profile_id"`
}

var (
	_ resource.Resource                = &assignmentGroupProfileBindingResource{}
	_ resource.ResourceWithConfigure   = &assignmentGroupProfileBindingResource{}
	_ resource.ResourceWithImportState = &assignmentGroupProfileBindingResource{}
)

func AssignmentGroupProfileBindingResource() resource.Resource {
	return &assignmentGroupProfileBindingResource{}
}

func (r *assignmentGroupProfileBindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assignmentgroup_profile_binding"
}

func (r *assignmentGroupProfileBindingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Binds an existing SimpleMDM profile (or custom declaration) to an existing assignment group. " +
			"Useful for managing memberships of UI-owned assignment groups (including dynamic groups) without " +
			"taking ownership of the assignment group itself.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<assignment_group_id>:<profile_id>\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"assignment_group_id": schema.StringAttribute{
				Required:    true,
				Description: "Identifier of the assignment group that should receive the profile.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"profile_id": schema.StringAttribute{
				Required: true,
				Description: "Identifier of the profile (or custom declaration) to bind. SimpleMDM's " +
					"/api/v1/profiles endpoint enumerates both profiles and custom declarations, so a custom " +
					"declaration ID works here too.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *assignmentGroupProfileBindingResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*simplemdm.Client)
}

func (r *assignmentGroupProfileBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan assignmentGroupProfileBindingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/profiles/%s",
		r.client.HostName,
		plan.AssignmentGroupID.ValueString(),
		plan.ProfileID.ValueString(),
	)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error building SimpleMDM profile-binding request", err.Error())
		return
	}

	if _, err := r.client.RequestResponse204or409(httpReq); err != nil {
		resp.Diagnostics.AddError("Error binding profile to assignment group", err.Error())
		return
	}

	// SimpleMDM has a short eventual-consistency window between when the
	// POST returns 204 and when /api/v1/profiles/{id} reflects the new
	// binding in its inverse `relationships.groups` array. The framework
	// runs a refresh immediately after Create and reports a non-empty plan
	// if Read decides the binding is gone. Block briefly until the binding
	// shows up so the post-Create refresh sees a consistent view.
	if err := r.waitForBindingVisible(ctx, plan.AssignmentGroupID.ValueString(), plan.ProfileID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Profile binding never appeared via inverse lookup", err.Error())
		return
	}

	plan.ID = types.StringValue(buildAssignmentGroupBindingID(plan.AssignmentGroupID.ValueString(), plan.ProfileID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// waitForBindingVisible polls /profiles/{id} until the inverse-lookup shows
// the assignment group, or until the budget is exhausted. SimpleMDM has a
// brief eventual-consistency window between the binding POST returning 204
// and /profiles/{id} reflecting the new binding in its inverse `groups`
// array; the wait keeps Create's post-step refresh from seeing transient
// drift. Empirically the binding is visible within a second on every probe
// we ran, so the generous ~15s budget is a safety margin rather than a
// commonly-hit ceiling.
func (r *assignmentGroupProfileBindingResource) waitForBindingVisible(ctx context.Context, agID, profileID string) error {
	delays := []time.Duration{0, 500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	url := fmt.Sprintf("https://%s/api/v1/profiles/%s", r.client.HostName, profileID)
	var lastErr error
	for _, d := range delays {
		if d > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(d):
			}
		}
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		body, err := r.client.RequestResponse200(httpReq)
		if err != nil {
			lastErr = err
			continue
		}
		bound, err := profileHasGroupAssignment(body, agID, profileID)
		if err != nil {
			lastErr = err
			continue
		}
		if bound {
			return nil
		}
		lastErr = fmt.Errorf("binding not yet visible")
	}
	return lastErr
}

func (r *assignmentGroupProfileBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state assignmentGroupProfileBindingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Strategy: fetch the profile and look for the assignment group in the
	// `relationships.groups` array. SimpleMDM exposes /api/v1/profiles/{id}
	// (unlike /custom_declarations/{id}, which 404s), and the response
	// includes the inverse `groups` relationship for both regular profiles
	// and custom declarations. Doing the inverse lookup here keeps Read O(1)
	// regardless of how many profiles the assignment group references.
	url := fmt.Sprintf("https://%s/api/v1/profiles/%s", r.client.HostName, state.ProfileID.ValueString())
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error building SimpleMDM profile request", err.Error())
		return
	}

	body, err := r.client.RequestResponse200(httpReq)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading SimpleMDM profile", err.Error())
		return
	}

	bound, err := profileHasGroupAssignment(body, state.AssignmentGroupID.ValueString(), state.ProfileID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error parsing SimpleMDM profile relationships", err.Error())
		return
	}
	if !bound {
		// Drift: someone removed the binding outside Terraform. Drop it from
		// state so the next apply recreates it (both attributes are ForceNew).
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(buildAssignmentGroupBindingID(state.AssignmentGroupID.ValueString(), state.ProfileID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *assignmentGroupProfileBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Both schema fields are ForceNew, so the framework will never call
	// Update with a non-trivial diff. Mirror the customDeclaration device
	// assignment resource and pass plan through to state so the framework
	// is happy if it ever does invoke us (e.g. for id recomputation).
	var plan assignmentGroupProfileBindingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(buildAssignmentGroupBindingID(plan.AssignmentGroupID.ValueString(), plan.ProfileID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *assignmentGroupProfileBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state assignmentGroupProfileBindingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://%s/api/v1/assignment_groups/%s/profiles/%s",
		r.client.HostName,
		state.AssignmentGroupID.ValueString(),
		state.ProfileID.ValueString(),
	)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error building SimpleMDM profile-binding delete request", err.Error())
		return
	}

	if _, err := r.client.RequestResponse204or409(httpReq); err != nil {
		if isNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Error removing profile binding", err.Error())
	}
}

func (r *assignmentGroupProfileBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	agID, profileID, err := parseAssignmentGroupBindingID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected import identifier format",
			"Expected assignment_group_id:profile_id or assignment_group_id|profile_id, got: "+req.ID,
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("assignment_group_id"), agID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("profile_id"), profileID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), buildAssignmentGroupBindingID(agID, profileID))...)
}

// profileHasGroupAssignment inspects a /api/v1/profiles/{id} response body
// and reports whether the profile's `relationships.groups` array includes
// the given assignment group. The group `id` in the relationships array is a
// JSON number; we decode with json.Number to avoid float64 scientific-notation
// truncation on IDs > 1e6 (which would silently miss real bindings).
func profileHasGroupAssignment(body []byte, assignmentGroupID, profileID string) (bool, error) {
	type relRef struct {
		ID json.Number `json:"id"`
	}
	type payload struct {
		Data struct {
			Relationships struct {
				Groups struct {
					Data []relRef `json:"data"`
				} `json:"groups"`
			} `json:"relationships"`
		} `json:"data"`
	}

	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		return false, fmt.Errorf("error parsing profile payload for profile %s: %w", profileID, err)
	}

	for _, g := range p.Data.Relationships.Groups.Data {
		if g.ID.String() == assignmentGroupID {
			return true, nil
		}
	}
	return false, nil
}

// buildAssignmentGroupBindingID returns the composite resource ID. Mirrors
// the customDeclaration_device_assignment convention: prefer ':' but fall
// back to '|' when an underlying ID already contains ':'.
func buildAssignmentGroupBindingID(assignmentGroupID, otherID string) string {
	if strings.Contains(assignmentGroupID, ":") || strings.Contains(otherID, ":") {
		return fmt.Sprintf("%s|%s", assignmentGroupID, otherID)
	}
	return fmt.Sprintf("%s:%s", assignmentGroupID, otherID)
}

// parseAssignmentGroupBindingID splits a composite resource ID into its two
// components. Accepts both ':' and '|' as separators for parity with the
// customDeclaration_device_assignment import behaviour.
func parseAssignmentGroupBindingID(raw string) (string, string, error) {
	var parts []string
	if strings.Contains(raw, "|") {
		parts = strings.Split(raw, "|")
	} else {
		parts = strings.Split(raw, ":")
	}
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid composite id: %q", raw)
	}
	return parts[0], parts[1], nil
}
