package provider

import (
	"context"
	"fmt"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &depServersDataSource{}
	_ datasource.DataSourceWithConfigure = &depServersDataSource{}
)

type depServersDataSource struct {
	client *simplemdm.Client
}

type depServersDataSourceModel struct {
	DepServers []depServerModel `tfsdk:"dep_servers"`
}

type depServerModel struct {
	ID               types.String `tfsdk:"id"`
	ServerName       types.String `tfsdk:"server_name"`
	ServerUUID       types.String `tfsdk:"server_uuid"`
	OrganizationName types.String `tfsdk:"organization_name"`
	DevicesFetchedAt types.String `tfsdk:"devices_fetched_at"`
	TokenExpiresAt   types.String `tfsdk:"token_expires_at"`
	LastSyncedAt     types.String `tfsdk:"last_synced_at"`
}

func DepServersDataSource() datasource.DataSource {
	return &depServersDataSource{}
}

func (d *depServersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dep_servers"
}

func (d *depServersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Apple DEP (Device Enrollment Program) servers associated with your SimpleMDM account.",
		Blocks: map[string]schema.Block{
			"dep_servers": schema.ListNestedBlock{
				Description: "Collection of DEP server records.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the DEP server.",
						},
						"server_name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the DEP server.",
						},
						"server_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "The UUID of the DEP server from Apple Business Manager.",
						},
						"organization_name": schema.StringAttribute{
							Computed:    true,
							Description: "The organization name from Apple Business Manager.",
						},
						"devices_fetched_at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when devices were last fetched from Apple.",
						},
						"token_expires_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the DEP token expires.",
						},
						"last_synced_at": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp of the most recent sync between SimpleMDM and Apple Business Manager.",
						},
					},
				},
			},
		},
	}
}

func (d *depServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state depServersDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	servers, err := simplemdmext.ListDepServers(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SimpleMDM DEP Servers",
			err.Error(),
		)
		return
	}

	state.DepServers = make([]depServerModel, 0, len(servers))
	for _, s := range servers {
		state.DepServers = append(state.DepServers, depServerModel{
			ID:               types.StringValue(fmt.Sprintf("%d", s.ID)),
			ServerName:       types.StringValue(s.Attributes.ServerName),
			ServerUUID:       types.StringValue(s.Attributes.ServerUUID),
			OrganizationName: types.StringValue(s.Attributes.OrganizationName),
			DevicesFetchedAt: types.StringValue(s.Attributes.DevicesFetchedAt),
			TokenExpiresAt:   stringValueOrNull(s.Attributes.TokenExpiresAt),
			LastSyncedAt:     stringValueOrNull(s.Attributes.LastSyncedAt),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *depServersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*simplemdm.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *simplemdm.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}
