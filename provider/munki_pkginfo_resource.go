package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &munkiPkgInfoResource{}
	_ resource.ResourceWithConfigure   = &munkiPkgInfoResource{}
	_ resource.ResourceWithImportState = &munkiPkgInfoResource{}
)

type munkiPkgInfoResource struct {
	client *simplemdm.Client
}

type munkiPkgInfoResourceModel struct {
	ID         types.String `tfsdk:"id"`
	AppID      types.String `tfsdk:"app_id"`
	Filename   types.String `tfsdk:"filename"`
	PkgInfo    types.String `tfsdk:"pkginfo"`
	PkgInfoSHA types.String `tfsdk:"pkginfo_sha256"`
}

func MunkiPkgInfoResource() resource.Resource {
	return &munkiPkgInfoResource{}
}

func (r *munkiPkgInfoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_munki_pkginfo"
}

func (r *munkiPkgInfoResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Munki pkginfo XML/PLIST blob attached to a Munki app. Apply uploads the supplied file via POST; Delete clears it via DELETE.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Composite identifier `munki_pkginfo:<app_id>`.",
			},
			"app_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Required. App ID to attach the pkginfo to.",
			},
			"filename": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional file name to send (default: `munki_pkginfo.plist`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pkginfo": schema.StringAttribute{
				Required:    true,
				Description: "Required. Pkginfo XML/PLIST contents (the body of the file).",
			},
			"pkginfo_sha256": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "SHA-256 fingerprint of the supplied pkginfo content (used to detect drift).",
			},
		},
	}
}

func (r *munkiPkgInfoResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResourceClient(req, resp)
}

func (r *munkiPkgInfoResource) apply(ctx context.Context, plan *munkiPkgInfoResourceModel, diags *resource.CreateResponse) bool {
	filename := plan.Filename.ValueString()
	if filename == "" {
		filename = "munki_pkginfo.plist"
	}
	content := []byte(plan.PkgInfo.ValueString())
	if err := simplemdmext.UploadMunkiPkgInfo(ctx, r.client, plan.AppID.ValueString(), content, filename); err != nil {
		diags.Diagnostics.AddError("Failed to upload Munki pkginfo", err.Error())
		return false
	}
	sum := sha256.Sum256(content)
	plan.ID = types.StringValue("munki_pkginfo:" + plan.AppID.ValueString())
	plan.Filename = types.StringValue(filename)
	plan.PkgInfoSHA = types.StringValue(hex.EncodeToString(sum[:]))
	return true
}

func (r *munkiPkgInfoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan munkiPkgInfoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !r.apply(ctx, &plan, resp) {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read is a no-op: the API does not expose a GET endpoint for pkginfo content.
func (r *munkiPkgInfoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state munkiPkgInfoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *munkiPkgInfoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan munkiPkgInfoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cr := &resource.CreateResponse{Diagnostics: resp.Diagnostics}
	if !r.apply(ctx, &plan, cr) {
		resp.Diagnostics = cr.Diagnostics
		return
	}
	resp.Diagnostics = cr.Diagnostics
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *munkiPkgInfoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state munkiPkgInfoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := simplemdmext.DeleteMunkiPkgInfo(ctx, r.client, state.AppID.ValueString()); err != nil {
		// Some apps don't support pkginfo deletion - warn but don't fail.
		resp.Diagnostics.AddWarning("Failed to delete Munki pkginfo", err.Error())
	}
}

// ImportState parses the composite id `munki_pkginfo:<app_id>` so importing
// also populates `app_id` (without it the resource would show drift on the
// next plan). The pkginfo body itself is unreadable via the API, so it stays
// unknown after import — first apply will re-upload from the user's config.
func (r *munkiPkgInfoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	appID := strings.TrimPrefix(id, "munki_pkginfo:")
	if appID == id || appID == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected `munki_pkginfo:<app_id>` for simplemdm_munki_pkginfo (the same value the resource produces in its id attribute).",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), appID)...)
}
