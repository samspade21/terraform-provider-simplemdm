package provider

import (
	"context"
	"fmt"
	"regexp"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &appInstallsDataSource{}
	_ datasource.DataSourceWithConfigure = &appInstallsDataSource{}
)

type appInstallsDataSource struct {
	client *simplemdm.Client
}

type appInstallsDataSourceModel struct {
	AppID    types.String      `tfsdk:"app_id"`
	Installs []appInstallModel `tfsdk:"installs"`
}

type appInstallModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Identifier   types.String `tfsdk:"identifier"`
	Version      types.String `tfsdk:"version"`
	ShortVersion types.String `tfsdk:"short_version"`
	BundleSize   types.Int64  `tfsdk:"bundle_size"`
	DynamicSize  types.Int64  `tfsdk:"dynamic_size"`
	Managed      types.Bool   `tfsdk:"managed"`
	DiscoveredAt types.String `tfsdk:"discovered_at"`
	LastSeenAt   types.String `tfsdk:"last_seen_at"`
	DeviceID     types.String `tfsdk:"device_id"`
}

func AppInstallsDataSource() datasource.DataSource {
	return &appInstallsDataSource{}
}

func (d *appInstallsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_installs"
}

func (d *appInstallsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every installation record (per device) of a managed application.",
		Attributes: map[string]schema.Attribute{
			"app_id": schema.StringAttribute{
				Required:    true,
				Description: "Required. ID of the application whose install records to list.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d+$`),
						"app_id must be a numeric string",
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"installs": schema.ListNestedBlock{
				Description: "Collection of install records for the application.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true, Description: "Installed app record ID."},
						"name":          schema.StringAttribute{Computed: true, Description: "Application name."},
						"identifier":    schema.StringAttribute{Computed: true, Description: "Application bundle identifier."},
						"version":       schema.StringAttribute{Computed: true, Description: "Application version."},
						"short_version": schema.StringAttribute{Computed: true, Description: "Application short version string."},
						"bundle_size":   schema.Int64Attribute{Computed: true, Description: "Bundle size in bytes."},
						"dynamic_size":  schema.Int64Attribute{Computed: true, Description: "Dynamic size in bytes."},
						"managed":       schema.BoolAttribute{Computed: true, Description: "Whether the install is managed by SimpleMDM."},
						"discovered_at": schema.StringAttribute{Computed: true, Description: "Timestamp when the install was first discovered."},
						"last_seen_at":  schema.StringAttribute{Computed: true, Description: "Timestamp when the install was last reported."},
						"device_id":     schema.StringAttribute{Computed: true, Description: "ID of the device the application is installed on."},
					},
				},
			},
		},
	}
}

func (d *appInstallsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state appInstallsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	installs, err := simplemdmext.ListAppInstalls(ctx, d.client, state.AppID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.Diagnostics.AddError("App not found",
				fmt.Sprintf("The app with ID %s was not found.", state.AppID.ValueString()))
			return
		}
		resp.Diagnostics.AddError("Unable to list app installs", err.Error())
		return
	}

	state.Installs = make([]appInstallModel, 0, len(installs))
	for _, item := range installs {
		row := appInstallModel{
			ID: types.StringValue(item.ID.String()),
		}

		if v, ok := item.Attributes["name"].(string); ok {
			row.Name = types.StringValue(v)
		} else {
			row.Name = types.StringNull()
		}
		if v, ok := item.Attributes["identifier"].(string); ok {
			row.Identifier = types.StringValue(v)
		} else {
			row.Identifier = types.StringNull()
		}
		if v, ok := item.Attributes["version"].(string); ok {
			row.Version = types.StringValue(v)
		} else {
			row.Version = types.StringNull()
		}
		if v, ok := item.Attributes["short_version"].(string); ok {
			row.ShortVersion = types.StringValue(v)
		} else {
			row.ShortVersion = types.StringNull()
		}
		if v, ok := item.Attributes["bundle_size"].(float64); ok {
			row.BundleSize = types.Int64Value(int64(v))
		} else {
			row.BundleSize = types.Int64Null()
		}
		if v, ok := item.Attributes["dynamic_size"].(float64); ok {
			row.DynamicSize = types.Int64Value(int64(v))
		} else {
			row.DynamicSize = types.Int64Null()
		}
		if v, ok := item.Attributes["managed"].(bool); ok {
			row.Managed = types.BoolValue(v)
		} else {
			row.Managed = types.BoolNull()
		}
		if v, ok := item.Attributes["discovered_at"].(string); ok {
			row.DiscoveredAt = types.StringValue(v)
		} else {
			row.DiscoveredAt = types.StringNull()
		}
		if v, ok := item.Attributes["last_seen_at"].(string); ok {
			row.LastSeenAt = types.StringValue(v)
		} else {
			row.LastSeenAt = types.StringNull()
		}
		if devID := item.Relationships.Device.Data.ID.String(); devID != "" {
			row.DeviceID = types.StringValue(devID)
		} else {
			row.DeviceID = types.StringNull()
		}

		state.Installs = append(state.Installs, row)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *appInstallsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
