package provider

import (
	"context"
	"time"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &deviceCustomAttributeValuesResource{}
	_ resource.ResourceWithConfigure = &deviceCustomAttributeValuesResource{}
)

type deviceCustomAttributeValuesResource struct {
	client *simplemdm.Client
}

type deviceAttributeValueAssignmentModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type deviceCustomAttributeValuesResourceModel struct {
	ID          types.String                          `tfsdk:"id"`
	DeviceID    types.String                          `tfsdk:"device_id"`
	Assignments []deviceAttributeValueAssignmentModel `tfsdk:"assignments"`
	Triggers    types.Map                             `tfsdk:"triggers"`
	LastApplied types.String                          `tfsdk:"last_applied"`
}

func DeviceCustomAttributeValuesResource() resource.Resource {
	return &deviceCustomAttributeValuesResource{}
}

func (r *deviceCustomAttributeValuesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_custom_attribute_values"
}

func (r *deviceCustomAttributeValuesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sets multiple custom attribute values on a single device in one PUT call. This is a fire-and-apply resource: changes to `assignments` apply on Update; the resource has no per-attribute drift detection. Use `triggers` to force a re-apply.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Synthetic identifier (`device_cav:<device_id>`).",
			},
			"device_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required. Device ID.",
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
			},
		},
		Blocks: map[string]schema.Block{
			"assignments": schema.ListNestedBlock{
				Description: "List of (attribute name, value) pairs to apply.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name":  schema.StringAttribute{Required: true, Description: "Custom attribute name (must already exist)."},
						"value": schema.StringAttribute{Required: true, Description: "Value to set for the attribute on this device."},
					},
				},
			},
		},
	}
}

func (r *deviceCustomAttributeValuesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

func (r *deviceCustomAttributeValuesResource) apply(ctx context.Context, plan *deviceCustomAttributeValuesResourceModel) error {
	assignments := make([]simplemdmext.DeviceAttributeAssignment, 0, len(plan.Assignments))
	for _, a := range plan.Assignments {
		assignments = append(assignments, simplemdmext.DeviceAttributeAssignment{
			Name:  a.Name.ValueString(),
			Value: a.Value.ValueString(),
		})
	}
	if err := simplemdmext.SetDeviceCustomAttributeValues(ctx, r.client, plan.DeviceID.ValueString(), assignments); err != nil {
		return err
	}
	plan.ID = types.StringValue("device_cav:" + plan.DeviceID.ValueString())
	plan.LastApplied = types.StringValue(time.Now().UTC().Format(time.RFC3339))
	return nil
}

func (r *deviceCustomAttributeValuesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceCustomAttributeValuesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to set device custom attribute values", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceCustomAttributeValuesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceCustomAttributeValuesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *deviceCustomAttributeValuesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deviceCustomAttributeValuesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Failed to update device custom attribute values", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceCustomAttributeValuesResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
