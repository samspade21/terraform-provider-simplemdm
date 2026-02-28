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
	_ datasource.DataSource              = &pushCertificateDataSource{}
	_ datasource.DataSourceWithConfigure = &pushCertificateDataSource{}
)

type pushCertificateDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	AppleID   types.String `tfsdk:"apple_id"`
	ExpiresAt types.String `tfsdk:"expires_at"`
	Subject   types.String `tfsdk:"subject"`
}

func PushCertificateDataSource() datasource.DataSource {
	return &pushCertificateDataSource{}
}

type pushCertificateDataSource struct {
	client *simplemdm.Client
}

func (d *pushCertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_push_certificate"
}

func (d *pushCertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about the Apple Push Notification Service (APNs) certificate for your SimpleMDM account. This is useful for monitoring certificate expiration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A static identifier for this data source (always 'push_certificate').",
			},
			"apple_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Apple ID associated with the push certificate.",
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "The expiration date and time of the push certificate.",
			},
			"subject": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate subject string.",
			},
		},
	}
}

func (d *pushCertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pushCertificateDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := simplemdmext.GetPushCertificate(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SimpleMDM Push Certificate",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue("push_certificate")
	state.AppleID = types.StringValue(cert.Data.Attributes.AppleID)
	state.ExpiresAt = types.StringValue(cert.Data.Attributes.ExpiresAt)
	state.Subject = types.StringValue(cert.Data.Attributes.Subject)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *pushCertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
