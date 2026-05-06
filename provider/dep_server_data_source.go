package provider

import (
	"context"
	"fmt"

	"github.com/DavidKrau/simplemdm-go-client"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &depServerDataSource{}
	_ datasource.DataSourceWithConfigure = &depServerDataSource{}
)

type depServerDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	ServerName       types.String `tfsdk:"server_name"`
	ServerUUID       types.String `tfsdk:"server_uuid"`
	OrganizationName types.String `tfsdk:"organization_name"`
	DevicesFetchedAt types.String `tfsdk:"devices_fetched_at"`
	TokenExpiresAt   types.String `tfsdk:"token_expires_at"`
	LastSyncedAt     types.String `tfsdk:"last_synced_at"`
}

func DepServerDataSource() datasource.DataSource {
	return &depServerDataSource{}
}

type depServerDataSource struct {
	client *simplemdm.Client
}

func (d *depServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dep_server"
}

func (d *depServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a specific Apple DEP (Device Enrollment Program) server associated with your SimpleMDM account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
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
				Description: "Timestamp when the DEP token expires (token must be renewed in Apple Business Manager before this date).",
			},
			"last_synced_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the most recent sync between SimpleMDM and Apple Business Manager.",
			},
		},
	}
}

func (d *depServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state depServerDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := simplemdmext.GetDepServer(ctx, d.client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SimpleMDM DEP Server",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d", server.Data.ID))
	state.ServerName = types.StringValue(server.Data.Attributes.ServerName)
	state.ServerUUID = types.StringValue(server.Data.Attributes.ServerUUID)
	state.OrganizationName = types.StringValue(server.Data.Attributes.OrganizationName)
	state.DevicesFetchedAt = types.StringValue(server.Data.Attributes.DevicesFetchedAt)
	state.TokenExpiresAt = stringValueOrNull(server.Data.Attributes.TokenExpiresAt)
	state.LastSyncedAt = stringValueOrNull(server.Data.Attributes.LastSyncedAt)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *depServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
