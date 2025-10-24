package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DavidKrau/simplemdm-go-client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &assignmentGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &assignmentGroupDataSource{}
)

type assignmentGroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	AutoDeploy  types.Bool   `tfsdk:"auto_deploy"`
	GroupType   types.String `tfsdk:"group_type"`
	InstallType types.String `tfsdk:"install_type"`
	Apps        types.Set    `tfsdk:"apps"`
	Groups      types.Set    `tfsdk:"groups"`
	Devices     types.Set    `tfsdk:"devices"`
	Profiles    types.Set    `tfsdk:"profiles"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	DeviceCount types.Int64  `tfsdk:"device_count"`
	GroupCount  types.Int64  `tfsdk:"group_count"`
}

func AssignmentGroupDataSource() datasource.DataSource {
	return &assignmentGroupDataSource{}
}

type assignmentGroupDataSource struct {
	client *simplemdm.Client
}

func (d *assignmentGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assignmentgroup"
}

func (d *assignmentGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Assignment Group data source exposes read-only information about existing assignment groups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the assignment group.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the assignment group.",
			},
			"auto_deploy": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the assignment group automatically deploys apps.",
			},
			"group_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of assignment group (standard or munki).",
			},
			"install_type": schema.StringAttribute{
				Computed:    true,
				Description: "Install type used when the assignment group is of type munki.",
			},
			"apps": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "IDs of apps assigned to the assignment group.",
			},
			"groups": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "IDs of device groups assigned to the assignment group.",
			},
			"devices": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "IDs of devices assigned directly to the assignment group.",
			},
			"profiles": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "IDs of profiles assigned to the assignment group.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the assignment group was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the assignment group was last updated.",
			},
			"device_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of devices currently assigned to the assignment group.",
			},
			"group_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of device groups currently assigned to the assignment group.",
			},
		},
	}
}

func (d *assignmentGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*simplemdm.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *assignmentGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state assignmentGroupDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assignmentGroup, err := fetchAssignmentGroup(ctx, d.client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SimpleMDM assignment group",
			err.Error(),
		)
		return
	}

	state.Name = types.StringValue(assignmentGroup.Data.Attributes.Name)
	state.AutoDeploy = types.BoolValue(assignmentGroup.Data.Attributes.AutoDeploy)
	state.GroupType = types.StringValue(assignmentGroup.Data.Attributes.Type)

	if assignmentGroup.Data.Attributes.Type == "munki" {
		state.InstallType = types.StringValue(assignmentGroup.Data.Attributes.InstallType)
	} else {
		state.InstallType = types.StringNull()
	}

	state.Apps = buildStringSetFromRelationshipItems(assignmentGroup.Data.Relationships.Apps.Data)
	state.Groups = buildStringSetFromRelationshipItems(assignmentGroup.Data.Relationships.DeviceGroups.Data)
	state.Devices = buildStringSetFromRelationshipItems(assignmentGroup.Data.Relationships.Devices.Data)
	state.Profiles = buildStringSetFromRelationshipItems(assignmentGroup.Data.Relationships.Profiles.Data)

	if assignmentGroup.Data.Attributes.CreatedAt != "" {
		state.CreatedAt = types.StringValue(assignmentGroup.Data.Attributes.CreatedAt)
	} else {
		state.CreatedAt = types.StringNull()
	}

	if assignmentGroup.Data.Attributes.UpdatedAt != "" {
		state.UpdatedAt = types.StringValue(assignmentGroup.Data.Attributes.UpdatedAt)
	} else {
		state.UpdatedAt = types.StringNull()
	}

	state.DeviceCount = types.Int64Value(int64(assignmentGroup.Data.Attributes.DeviceCount))
	state.GroupCount = types.Int64Value(int64(assignmentGroup.Data.Attributes.GroupCount))

	state.ID = types.StringValue(strconv.Itoa(assignmentGroup.Data.ID))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
