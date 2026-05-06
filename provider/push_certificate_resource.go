package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	simplemdm "github.com/DavidKrau/simplemdm-go-client"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &pushCertificateResource{}
	_ resource.ResourceWithConfigure = &pushCertificateResource{}
)

type pushCertificateResource struct {
	client *simplemdm.Client
}

type pushCertificateResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Certificate       types.String `tfsdk:"certificate"`
	CertificateSHA256 types.String `tfsdk:"certificate_sha256"`
	AppleID           types.String `tfsdk:"apple_id"`
	ExpiresAt         types.String `tfsdk:"expires_at"`
}

func PushCertificateResource() resource.Resource {
	return &pushCertificateResource{}
}

func (r *pushCertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_push_certificate"
}

func (r *pushCertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the SimpleMDM tenant's Apple Push Notification certificate. Apply uploads the supplied PEM bytes via PUT /push_certificate. Deleting the resource only removes it from Terraform state; it does NOT clear the certificate on the SimpleMDM side.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Synthetic identifier (`push_certificate`).",
			},
			"certificate": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Required. The push certificate as raw PEM bytes (string). Send the contents of the .pem file Apple provides.",
			},
			"certificate_sha256": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "SHA-256 fingerprint of the supplied certificate, used to detect drift.",
			},
			"apple_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional. Email address of the Apple ID the certificate was generated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "Certificate expiry timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *pushCertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

func (r *pushCertificateResource) apply(ctx context.Context, plan *pushCertificateResourceModel, diags *diag.Diagnostics) {
	cert := []byte(plan.Certificate.ValueString())
	uploaded, err := simplemdmext.UploadPushCertificate(ctx, r.client, cert, plan.AppleID.ValueString())
	if err != nil {
		diags.AddError("Failed to upload SimpleMDM push certificate", err.Error())
		return
	}
	sum := sha256.Sum256(cert)
	plan.ID = types.StringValue("push_certificate")
	plan.CertificateSHA256 = types.StringValue(hex.EncodeToString(sum[:]))
	plan.AppleID = types.StringValue(uploaded.Data.Attributes.AppleID)
	plan.ExpiresAt = types.StringValue(uploaded.Data.Attributes.ExpiresAt)
}

func (r *pushCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pushCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pushCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pushCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := simplemdmext.GetPushCertificate(ctx, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SimpleMDM push certificate", err.Error())
		return
	}
	state.AppleID = types.StringValue(current.Data.Attributes.AppleID)
	state.ExpiresAt = types.StringValue(current.Data.Attributes.ExpiresAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pushCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pushCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.apply(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op: SimpleMDM does not support clearing the push certificate
// via API. State removal only.
func (r *pushCertificateResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
