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
	_ datasource.DataSource              = &depDeviceDataSource{}
	_ datasource.DataSourceWithConfigure = &depDeviceDataSource{}
)

type depDeviceDataSource struct {
	client *simplemdm.Client
}

type depDeviceDataSourceModel struct {
	DepServerID        types.String `tfsdk:"dep_server_id"`
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

func DepDeviceDataSource() datasource.DataSource {
	return &depDeviceDataSource{}
}

func (d *depDeviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dep_device"
}

func (d *depDeviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a single DEP device under a DEP server by its DEP device ID.",
		Attributes: map[string]schema.Attribute{
			"dep_server_id": schema.StringAttribute{
				Required:    true,
				Description: "Required. The DEP server ID this device belongs to.",
			},
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Required. DEP device ID.",
			},
			"serial_number":        schema.StringAttribute{Computed: true},
			"model":                schema.StringAttribute{Computed: true},
			"color":                schema.StringAttribute{Computed: true},
			"description":          schema.StringAttribute{Computed: true},
			"os":                   schema.StringAttribute{Computed: true},
			"device_family":        schema.StringAttribute{Computed: true},
			"profile_status":       schema.StringAttribute{Computed: true},
			"profile_assign_time":  schema.StringAttribute{Computed: true},
			"profile_push_time":    schema.StringAttribute{Computed: true},
			"device_assigned_date": schema.StringAttribute{Computed: true},
			"device_assigned_by":   schema.StringAttribute{Computed: true},
		},
	}
}

func (d *depDeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state depDeviceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dev, err := simplemdmext.GetDepDevice(ctx, d.client, state.DepServerID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read DEP device", err.Error())
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d", dev.Data.ID))
	state.SerialNumber = types.StringValue(dev.Data.Attributes.SerialNumber)
	state.Model = types.StringValue(dev.Data.Attributes.Model)
	state.Color = types.StringValue(dev.Data.Attributes.Color)
	state.Description = types.StringValue(dev.Data.Attributes.Description)
	state.OS = types.StringValue(dev.Data.Attributes.OS)
	state.DeviceFamily = types.StringValue(dev.Data.Attributes.DeviceFamily)
	state.ProfileStatus = types.StringValue(dev.Data.Attributes.ProfileStatus)
	state.ProfileAssignTime = types.StringValue(dev.Data.Attributes.ProfileAssignTime)
	state.ProfilePushTime = types.StringValue(dev.Data.Attributes.ProfilePushTime)
	state.DeviceAssignedDate = types.StringValue(dev.Data.Attributes.DeviceAssignedDate)
	state.DeviceAssignedBy = types.StringValue(dev.Data.Attributes.DeviceAssignedBy)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *depDeviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
