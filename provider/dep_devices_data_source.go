package provider

import (
	"context"
	"fmt"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &depDevicesDataSource{}
	_ datasource.DataSourceWithConfigure = &depDevicesDataSource{}
)

type depDevicesDataSource struct {
	client *simplemdm.Client
}

type depDevicesDataSourceModel struct {
	DepServerID types.String     `tfsdk:"dep_server_id"`
	DepDevices  []depDeviceModel `tfsdk:"dep_devices"`
}

type depDeviceModel struct {
	ID                 types.String `tfsdk:"id"`
	SerialNumber       types.String `tfsdk:"serial_number"`
	Model              types.String `tfsdk:"model"`
	Color              types.String `tfsdk:"color"`
	Description        types.String `tfsdk:"description"`
	OS                 types.String `tfsdk:"os"`
	DeviceFamily       types.String `tfsdk:"device_family"`
	ProfileStatus      types.String `tfsdk:"profile_status"`
	ProfileAssignTime  types.String `tfsdk:"profile_assign_time"`
	ProfilePushTime    types.String `tfsdk:"profile_push_time"`
	DeviceAssignedDate types.String `tfsdk:"device_assigned_date"`
	DeviceAssignedBy   types.String `tfsdk:"device_assigned_by"`
}

func DepDevicesDataSource() datasource.DataSource {
	return &depDevicesDataSource{}
}

func (d *depDevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dep_devices"
}

func (d *depDevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists DEP devices reported by the given Apple DEP server.",
		Attributes: map[string]schema.Attribute{
			"dep_server_id": schema.StringAttribute{
				Required:    true,
				Description: "Required. The DEP server ID whose devices to list.",
			},
		},
		Blocks: map[string]schema.Block{
			"dep_devices": schema.ListNestedBlock{
				Description: "Collection of DEP devices.",
				NestedObject: schema.NestedBlockObject{
					Attributes: depDeviceAttributesSchema(),
				},
			},
		},
	}
}

func depDeviceAttributesSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":                   schema.StringAttribute{Computed: true, Description: "DEP device ID."},
		"serial_number":        schema.StringAttribute{Computed: true, Description: "Serial number reported by Apple."},
		"model":                schema.StringAttribute{Computed: true, Description: "Apple device model."},
		"color":                schema.StringAttribute{Computed: true, Description: "Reported device color."},
		"description":          schema.StringAttribute{Computed: true, Description: "Reported device description."},
		"os":                   schema.StringAttribute{Computed: true, Description: "Operating system family."},
		"device_family":        schema.StringAttribute{Computed: true, Description: "Apple device family (iPhone, iPad, Mac…)."},
		"profile_status":       schema.StringAttribute{Computed: true, Description: "Current profile assignment status."},
		"profile_assign_time":  schema.StringAttribute{Computed: true, Description: "Timestamp when the device profile was assigned."},
		"profile_push_time":    schema.StringAttribute{Computed: true, Description: "Timestamp when the profile was last pushed."},
		"device_assigned_date": schema.StringAttribute{Computed: true, Description: "Date the device was assigned to the organization."},
		"device_assigned_by":   schema.StringAttribute{Computed: true, Description: "User or system that assigned the device."},
	}
}

func (d *depDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state depDevicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	devices, err := simplemdmext.ListDepDevices(ctx, d.client, state.DepServerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to list DEP devices", err.Error())
		return
	}

	state.DepDevices = make([]depDeviceModel, 0, len(devices))
	for _, dev := range devices {
		state.DepDevices = append(state.DepDevices, mapDepDevice(dev))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapDepDevice(d simplemdmext.DepDeviceData) depDeviceModel {
	return depDeviceModel{
		ID:                 types.StringValue(fmt.Sprintf("%d", d.ID)),
		SerialNumber:       types.StringValue(d.Attributes.SerialNumber),
		Model:              types.StringValue(d.Attributes.Model),
		Color:              types.StringValue(d.Attributes.Color),
		Description:        types.StringValue(d.Attributes.Description),
		OS:                 types.StringValue(d.Attributes.OS),
		DeviceFamily:       types.StringValue(d.Attributes.DeviceFamily),
		ProfileStatus:      types.StringValue(d.Attributes.ProfileStatus),
		ProfileAssignTime:  types.StringValue(d.Attributes.ProfileAssignTime),
		ProfilePushTime:    types.StringValue(d.Attributes.ProfilePushTime),
		DeviceAssignedDate: types.StringValue(d.Attributes.DeviceAssignedDate),
		DeviceAssignedBy:   types.StringValue(d.Attributes.DeviceAssignedBy),
	}
}

func (d *depDevicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*simplemdm.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *simplemdm.Client, got: %T.", req.ProviderData))
		return
	}
	d.client = client
}
