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
	_ datasource.DataSource              = &pushCertificateSCSRDataSource{}
	_ datasource.DataSourceWithConfigure = &pushCertificateSCSRDataSource{}
)

type pushCertificateSCSRDataSource struct {
	client *simplemdm.Client
}

type pushCertificateSCSRDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Data types.String `tfsdk:"data"`
}

func PushCertificateSCSRDataSource() datasource.DataSource {
	return &pushCertificateSCSRDataSource{}
}

func (d *pushCertificateSCSRDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_push_certificate_scsr"
}

func (d *pushCertificateSCSRDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a base64-encoded plist (signed CSR) for upload to the Apple Push Certificates Portal.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic identifier; always 'push_certificate_scsr'.",
			},
			"data": schema.StringAttribute{
				Computed:    true,
				Description: "Base64-encoded plist value to upload to Apple as-is.",
			},
		},
	}
}

func (d *pushCertificateSCSRDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp_, err := simplemdmext.GetPushCertificateSCSR(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read push certificate SCSR", err.Error())
		return
	}
	state := pushCertificateSCSRDataSourceModel{
		ID:   types.StringValue("push_certificate_scsr"),
		Data: types.StringValue(resp_.Data),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *pushCertificateSCSRDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
