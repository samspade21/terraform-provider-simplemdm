package provider

import (
	"context"
	"fmt"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// customAttributeValueModel is the shared per-item shape for the three "list
// custom attribute values" endpoints.
type customAttributeValueModel struct {
	ID     types.String `tfsdk:"id"`
	Value  types.String `tfsdk:"value"`
	Secret types.Bool   `tfsdk:"secret"`
	Source types.String `tfsdk:"source"`
}

func customAttributeValueAttributesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":     schema.StringAttribute{Computed: true, Description: "Custom attribute name."},
		"value":  schema.StringAttribute{Computed: true, Description: "Resolved attribute value."},
		"secret": schema.BoolAttribute{Computed: true, Description: "Whether the attribute is marked as secret."},
		"source": schema.StringAttribute{Computed: true, Description: "Source of the value: 'device', 'group', or 'account'."},
	}
}

func mapAttributeArray(arr *simplemdm.AttributeArray) []customAttributeValueModel {
	if arr == nil {
		return nil
	}
	out := make([]customAttributeValueModel, 0, len(arr.Data))
	for _, item := range arr.Data {
		out = append(out, customAttributeValueModel{
			ID:     types.StringValue(item.ID),
			Value:  types.StringValue(item.Attributes.Value),
			Secret: types.BoolValue(item.Attributes.Secret),
			Source: types.StringValue(item.Attributes.Source),
		})
	}
	return out
}

// ============================================================
// Device custom attribute values
// ============================================================

var (
	_ datasource.DataSource              = &deviceCustomAttributeValuesDataSource{}
	_ datasource.DataSourceWithConfigure = &deviceCustomAttributeValuesDataSource{}
)

type deviceCustomAttributeValuesDataSource struct {
	client *simplemdm.Client
}

type deviceCustomAttributeValuesModel struct {
	DeviceID              types.String                `tfsdk:"device_id"`
	CustomAttributeValues []customAttributeValueModel `tfsdk:"custom_attribute_values"`
}

func DeviceCustomAttributeValuesDataSource() datasource.DataSource {
	return &deviceCustomAttributeValuesDataSource{}
}

func (d *deviceCustomAttributeValuesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_custom_attribute_values"
}

func (d *deviceCustomAttributeValuesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists custom attribute values resolved for a device, including overrides from groups or the account default.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{Required: true, Description: "Required. ID of the device."},
		},
		Blocks: map[string]schema.Block{
			"custom_attribute_values": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{Attributes: customAttributeValueAttributesSchema()},
			},
		},
	}
}

func (d *deviceCustomAttributeValuesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceCustomAttributeValuesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	arr, err := d.client.AttributeGetAttributesForDevice(state.DeviceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read device custom attribute values", err.Error())
		return
	}
	state.CustomAttributeValues = mapAttributeArray(arr)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *deviceCustomAttributeValuesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSourceClient(req, resp)
}

// ============================================================
// Assignment group custom attribute values
// ============================================================

var (
	_ datasource.DataSource              = &assignmentGroupCustomAttributeValuesDataSource{}
	_ datasource.DataSourceWithConfigure = &assignmentGroupCustomAttributeValuesDataSource{}
)

type assignmentGroupCustomAttributeValuesDataSource struct {
	client *simplemdm.Client
}

type assignmentGroupCustomAttributeValuesModel struct {
	AssignmentGroupID     types.String                `tfsdk:"assignment_group_id"`
	CustomAttributeValues []customAttributeValueModel `tfsdk:"custom_attribute_values"`
}

func AssignmentGroupCustomAttributeValuesDataSource() datasource.DataSource {
	return &assignmentGroupCustomAttributeValuesDataSource{}
}

func (d *assignmentGroupCustomAttributeValuesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assignmentgroup_custom_attribute_values"
}

func (d *assignmentGroupCustomAttributeValuesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists custom attribute values assigned to an assignment group.",
		Attributes: map[string]schema.Attribute{
			"assignment_group_id": schema.StringAttribute{Required: true, Description: "Required. ID of the assignment group."},
		},
		Blocks: map[string]schema.Block{
			"custom_attribute_values": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{Attributes: customAttributeValueAttributesSchema()},
			},
		},
	}
}

func (d *assignmentGroupCustomAttributeValuesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state assignmentGroupCustomAttributeValuesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	arr, err := d.client.AttributeGetAttributesForGroup(state.AssignmentGroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read assignment group custom attribute values", err.Error())
		return
	}
	state.CustomAttributeValues = mapAttributeArray(arr)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *assignmentGroupCustomAttributeValuesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSourceClient(req, resp)
}

// ============================================================
// Device group custom attribute values
// ============================================================

var (
	_ datasource.DataSource              = &deviceGroupCustomAttributeValuesDataSource{}
	_ datasource.DataSourceWithConfigure = &deviceGroupCustomAttributeValuesDataSource{}
)

type deviceGroupCustomAttributeValuesDataSource struct {
	client *simplemdm.Client
}

type deviceGroupCustomAttributeValuesModel struct {
	DeviceGroupID         types.String                `tfsdk:"device_group_id"`
	CustomAttributeValues []customAttributeValueModel `tfsdk:"custom_attribute_values"`
}

func DeviceGroupCustomAttributeValuesDataSource() datasource.DataSource {
	return &deviceGroupCustomAttributeValuesDataSource{}
}

func (d *deviceGroupCustomAttributeValuesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devicegroup_custom_attribute_values"
}

func (d *deviceGroupCustomAttributeValuesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists custom attribute values assigned to a (legacy) device group.",
		Attributes: map[string]schema.Attribute{
			"device_group_id": schema.StringAttribute{Required: true, Description: "Required. ID of the device group."},
		},
		Blocks: map[string]schema.Block{
			"custom_attribute_values": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{Attributes: customAttributeValueAttributesSchema()},
			},
		},
	}
}

func (d *deviceGroupCustomAttributeValuesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceGroupCustomAttributeValuesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	arr, err := d.client.AttributeGetAttributesForDeviceGroup(state.DeviceGroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read device group custom attribute values", err.Error())
		return
	}
	state.CustomAttributeValues = mapAttributeArray(arr)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *deviceGroupCustomAttributeValuesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSourceClient(req, resp)
}

// ============================================================
// Helpers
// ============================================================

func configureDataSourceClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *simplemdm.Client {
	if req.ProviderData == nil {
		return nil
	}
	client, ok := req.ProviderData.(*simplemdm.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *simplemdm.Client, got: %T.", req.ProviderData))
		return nil
	}
	return client
}
