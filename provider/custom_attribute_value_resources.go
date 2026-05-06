package provider

import (
	"context"
	"fmt"
	"time"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// findAttributeValue searches an AttributeArray for the named attribute and
// returns its current value, or empty if not present.
func findAttributeValue(arr *simplemdm.AttributeArray, attrName string) (value string, found bool) {
	if arr == nil {
		return "", false
	}
	for _, item := range arr.Data {
		if item.ID == attrName {
			return item.Attributes.Value, true
		}
	}
	return "", false
}

func configureResourceClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *simplemdm.Client {
	if req.ProviderData == nil {
		return nil
	}
	client, ok := req.ProviderData.(*simplemdm.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *simplemdm.Client, got: %T.", req.ProviderData))
		return nil
	}
	return client
}

// ============================================================
// simplemdm_device_custom_attribute_value
// ============================================================

var (
	_ resource.Resource              = &deviceCustomAttributeValueResource{}
	_ resource.ResourceWithConfigure = &deviceCustomAttributeValueResource{}
)

type deviceCustomAttributeValueResource struct {
	client *simplemdm.Client
}

type deviceCustomAttributeValueModel struct {
	ID            types.String `tfsdk:"id"`
	DeviceID      types.String `tfsdk:"device_id"`
	AttributeName types.String `tfsdk:"attribute_name"`
	Value         types.String `tfsdk:"value"`
}

func DeviceCustomAttributeValueResource() resource.Resource {
	return &deviceCustomAttributeValueResource{}
}

func (r *deviceCustomAttributeValueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_custom_attribute_value"
}

func (r *deviceCustomAttributeValueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sets a custom attribute value on a specific device. Deleting the resource clears the value (sets it to empty string).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite ID `<device_id>/<attribute_name>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required. Device ID.",
			},
			"attribute_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required. Custom attribute name (must already exist).",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "Value to assign to the attribute on this device.",
			},
		},
	}
}

func (r *deviceCustomAttributeValueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

func (r *deviceCustomAttributeValueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceCustomAttributeValueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AttributeSetAttributeForDevice(plan.DeviceID.ValueString(), plan.AttributeName.ValueString(), plan.Value.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to set device custom attribute value", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.DeviceID.ValueString() + "/" + plan.AttributeName.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceCustomAttributeValueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceCustomAttributeValueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	arr, err := r.client.AttributeGetAttributesForDevice(state.DeviceID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read device custom attribute value", err.Error())
		return
	}
	if v, ok := findAttributeValue(arr, state.AttributeName.ValueString()); ok {
		state.Value = types.StringValue(v)
	} else {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *deviceCustomAttributeValueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deviceCustomAttributeValueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AttributeSetAttributeForDevice(plan.DeviceID.ValueString(), plan.AttributeName.ValueString(), plan.Value.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to update device custom attribute value", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.DeviceID.ValueString() + "/" + plan.AttributeName.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceCustomAttributeValueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceCustomAttributeValueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Best-effort: clear the value. SimpleMDM has no way to "delete" the
	// per-device override, but setting it to empty effectively unsets it.
	if err := r.client.AttributeSetAttributeForDevice(state.DeviceID.ValueString(), state.AttributeName.ValueString(), ""); err != nil {
		// Surface as warning; nothing the user can fix beyond network issues.
		resp.Diagnostics.AddWarning("Failed to clear device custom attribute value on delete", err.Error())
	}
}

// ============================================================
// simplemdm_assignmentgroup_custom_attribute_value
// ============================================================

var (
	_ resource.Resource              = &assignmentGroupCustomAttributeValueResource{}
	_ resource.ResourceWithConfigure = &assignmentGroupCustomAttributeValueResource{}
)

type assignmentGroupCustomAttributeValueResource struct {
	client *simplemdm.Client
}

type assignmentGroupCustomAttributeValueModel struct {
	ID                types.String `tfsdk:"id"`
	AssignmentGroupID types.String `tfsdk:"assignment_group_id"`
	AttributeName     types.String `tfsdk:"attribute_name"`
	Value             types.String `tfsdk:"value"`
}

func AssignmentGroupCustomAttributeValueResource() resource.Resource {
	return &assignmentGroupCustomAttributeValueResource{}
}

func (r *assignmentGroupCustomAttributeValueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assignmentgroup_custom_attribute_value"
}

func (r *assignmentGroupCustomAttributeValueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sets a custom attribute value on an assignment group. Deleting the resource clears the value.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Composite ID `<assignment_group_id>/<attribute_name>`.",
			},
			"assignment_group_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required. Assignment group ID.",
			},
			"attribute_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required. Custom attribute name (must already exist).",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "Value to assign to the attribute on this assignment group.",
			},
		},
	}
}

func (r *assignmentGroupCustomAttributeValueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

func (r *assignmentGroupCustomAttributeValueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan assignmentGroupCustomAttributeValueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AttributeSetAttributeForDeviceGroup(plan.AssignmentGroupID.ValueString(), plan.AttributeName.ValueString(), plan.Value.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to set assignment group custom attribute value", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.AssignmentGroupID.ValueString() + "/" + plan.AttributeName.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *assignmentGroupCustomAttributeValueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state assignmentGroupCustomAttributeValueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	arr, err := r.client.AttributeGetAttributesForGroup(state.AssignmentGroupID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read assignment group custom attribute value", err.Error())
		return
	}
	if v, ok := findAttributeValue(arr, state.AttributeName.ValueString()); ok {
		state.Value = types.StringValue(v)
	} else {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *assignmentGroupCustomAttributeValueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan assignmentGroupCustomAttributeValueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AttributeSetAttributeForDeviceGroup(plan.AssignmentGroupID.ValueString(), plan.AttributeName.ValueString(), plan.Value.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to update assignment group custom attribute value", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.AssignmentGroupID.ValueString() + "/" + plan.AttributeName.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *assignmentGroupCustomAttributeValueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state assignmentGroupCustomAttributeValueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AttributeSetAttributeForDeviceGroup(state.AssignmentGroupID.ValueString(), state.AttributeName.ValueString(), ""); err != nil {
		resp.Diagnostics.AddWarning("Failed to clear assignment group custom attribute value on delete", err.Error())
	}
}

// ============================================================
// simplemdm_custom_attribute_bulk_value (multi-device set)
// ============================================================

var (
	_ resource.Resource              = &customAttributeBulkValueResource{}
	_ resource.ResourceWithConfigure = &customAttributeBulkValueResource{}
)

type customAttributeBulkValueResource struct {
	client *simplemdm.Client
}

type customAttributeBulkAssignmentModel struct {
	DeviceID types.String `tfsdk:"device_id"`
	Value    types.String `tfsdk:"value"`
}

type customAttributeBulkValueModel struct {
	ID            types.String                         `tfsdk:"id"`
	AttributeName types.String                         `tfsdk:"attribute_name"`
	Assignments   []customAttributeBulkAssignmentModel `tfsdk:"assignments"`
	Triggers      types.Map                            `tfsdk:"triggers"`
	LastApplied   types.String                         `tfsdk:"last_applied"`
}

func CustomAttributeBulkValueResource() resource.Resource {
	return &customAttributeBulkValueResource{}
}

func (r *customAttributeBulkValueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_attribute_bulk_value"
}

func (r *customAttributeBulkValueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Pushes a custom attribute value across multiple devices in one PUT call. This is a fire-and-apply resource: changes to `assignments` apply on Update; the resource has no per-device drift detection. Use `triggers` to force a re-apply.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Synthetic identifier (`bulk:<attribute_name>`).",
			},
			"attribute_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required. Custom attribute name (must already exist).",
			},
			"triggers": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Optional arbitrary string map. Changing any value forces resource replacement, re-applying the bulk update.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"last_applied": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Timestamp of the most recent successful apply.",
			},
		},
		Blocks: map[string]schema.Block{
			"assignments": schema.ListNestedBlock{
				Description: "List of (device_id, value) pairs to apply.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"device_id": schema.StringAttribute{Required: true, Description: "Device ID."},
						"value":     schema.StringAttribute{Required: true, Description: "Value to assign on the device."},
					},
				},
			},
		},
	}
}

func (r *customAttributeBulkValueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

func (r *customAttributeBulkValueResource) apply(ctx context.Context, plan *customAttributeBulkValueModel, diagnostics *resource.CreateResponse) bool {
	assignments := make([]simplemdmext.BulkAttributeAssignment, 0, len(plan.Assignments))
	for _, a := range plan.Assignments {
		assignments = append(assignments, simplemdmext.BulkAttributeAssignment{
			DeviceID: a.DeviceID.ValueString(),
			Value:    a.Value.ValueString(),
		})
	}
	if err := simplemdmext.BulkSetCustomAttributeValue(ctx, r.client, plan.AttributeName.ValueString(), assignments); err != nil {
		diagnostics.Diagnostics.AddError("Failed to bulk-set custom attribute value", err.Error())
		return false
	}
	plan.ID = types.StringValue("bulk:" + plan.AttributeName.ValueString())
	plan.LastApplied = types.StringValue(time.Now().UTC().Format(time.RFC3339))
	return true
}

func (r *customAttributeBulkValueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customAttributeBulkValueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !r.apply(ctx, &plan, resp) {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customAttributeBulkValueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customAttributeBulkValueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customAttributeBulkValueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customAttributeBulkValueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Build a CreateResponse-shaped struct so we can reuse apply()'s Diagnostics.
	cr := &resource.CreateResponse{Diagnostics: resp.Diagnostics}
	if !r.apply(ctx, &plan, cr) {
		resp.Diagnostics = cr.Diagnostics
		return
	}
	resp.Diagnostics = cr.Diagnostics
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customAttributeBulkValueResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Removing a bulk-apply resource does not undo previous PUTs. State removal only.
}
