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
	_ datasource.DataSource              = &accountDataSource{}
	_ datasource.DataSourceWithConfigure = &accountDataSource{}
)

type accountDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	DefaultDeviceGroupID types.String `tfsdk:"default_device_group_id"`
	CarrierActivation    types.Bool   `tfsdk:"carrier_activation"`
	DepEnabled           types.Bool   `tfsdk:"dep_enabled"`
	AppUpdatesEnabled    types.Bool   `tfsdk:"app_updates_enabled"`
}

func AccountDataSource() datasource.DataSource {
	return &accountDataSource{}
}

type accountDataSource struct {
	client *simplemdm.Client
}

func (d *accountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *accountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves SimpleMDM account information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the SimpleMDM account.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The company name associated with the account.",
			},
			"default_device_group_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the default device group for newly enrolled devices.",
			},
			"carrier_activation": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether carrier activation is enabled for the account.",
			},
			"dep_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether Apple DEP (Device Enrollment Program) is enabled.",
			},
			"app_updates_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether automatic app updates are enabled.",
			},
		},
	}
}

func (d *accountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accountDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := simplemdmext.GetAccount(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SimpleMDM Account",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d", account.Data.ID))
	state.Name = types.StringValue(account.Data.Attributes.Name)
	state.CarrierActivation = types.BoolValue(account.Data.Attributes.CarrierActivation)
	state.DepEnabled = types.BoolValue(account.Data.Attributes.DepEnabled)
	state.AppUpdatesEnabled = types.BoolValue(account.Data.Attributes.AppUpdatesEnabled)

	if account.Data.Attributes.DefaultDeviceGroupID != nil {
		state.DefaultDeviceGroupID = types.StringValue(fmt.Sprintf("%d", *account.Data.Attributes.DefaultDeviceGroupID))
	} else {
		state.DefaultDeviceGroupID = types.StringValue("")
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *accountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
