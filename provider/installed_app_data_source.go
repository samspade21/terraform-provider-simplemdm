package provider

import (
	"context"
	"fmt"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &installedAppDataSource{}
	_ datasource.DataSourceWithConfigure = &installedAppDataSource{}
)

type installedAppDataSource struct {
	client *simplemdm.Client
}

type installedAppDataSourceModel struct {
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
}

func InstalledAppDataSource() datasource.DataSource {
	return &installedAppDataSource{}
}

func (d *installedAppDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_installed_app"
}

func (d *installedAppDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a single installed app record by ID.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Required: true, Description: "Required. Installed app ID."},
			"name":          schema.StringAttribute{Computed: true, Description: "Application name."},
			"identifier":    schema.StringAttribute{Computed: true, Description: "Application bundle identifier."},
			"version":       schema.StringAttribute{Computed: true, Description: "Application version."},
			"short_version": schema.StringAttribute{Computed: true, Description: "Application short version string."},
			"bundle_size":   schema.Int64Attribute{Computed: true, Description: "Bundle size in bytes."},
			"dynamic_size":  schema.Int64Attribute{Computed: true, Description: "Dynamic size in bytes."},
			"managed":       schema.BoolAttribute{Computed: true, Description: "Whether the app is managed by SimpleMDM."},
			"discovered_at": schema.StringAttribute{Computed: true},
			"last_seen_at":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *installedAppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state installedAppDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := simplemdmext.GetInstalledApp(ctx, d.client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read installed app", err.Error())
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d", app.Data.ID))
	state.Name = types.StringValue(app.Data.Attributes.Name)
	state.Identifier = types.StringValue(app.Data.Attributes.Identifier)
	state.Version = types.StringValue(app.Data.Attributes.Version)
	state.ShortVersion = types.StringValue(app.Data.Attributes.ShortVersion)
	state.BundleSize = types.Int64Value(app.Data.Attributes.BundleSize)
	if app.Data.Attributes.DynamicSize != nil {
		state.DynamicSize = types.Int64Value(*app.Data.Attributes.DynamicSize)
	} else {
		state.DynamicSize = types.Int64Null()
	}
	state.Managed = types.BoolValue(app.Data.Attributes.Managed)
	state.DiscoveredAt = types.StringValue(app.Data.Attributes.DiscoveredAt)
	state.LastSeenAt = types.StringValue(app.Data.Attributes.LastSeenAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *installedAppDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
