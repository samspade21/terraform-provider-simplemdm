package provider

import (
	"context"
	"fmt"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &accountResource{}
	_ resource.ResourceWithConfigure   = &accountResource{}
	_ resource.ResourceWithImportState = &accountResource{}
)

type accountResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	AppleStoreCountryCode types.String `tfsdk:"apple_store_country_code"`
}

func AccountResource() resource.Resource {
	return &accountResource{}
}

type accountResource struct {
	client *simplemdm.Client
}

func (r *accountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (r *accountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages SimpleMDM account-level settings (singleton). Creating this resource overwrites the existing tenant settings; deleting it only removes the resource from Terraform state and does not modify the tenant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Numeric SimpleMDM account ID.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Company name associated with the account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"apple_store_country_code": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Apple App Store country code (e.g. US, AU).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *accountResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*simplemdm.Client)
}

func (r *accountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyAndRefresh(ctx, plan.Name.ValueString(), plan.AppleStoreCountryCode.ValueString(), &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *accountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := simplemdmext.GetAccount(ctx, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading SimpleMDM Account", err.Error())
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d", account.Data.ID))
	state.Name = types.StringValue(account.Data.Attributes.Name)
	state.AppleStoreCountryCode = types.StringValue(account.Data.Attributes.AppleStoreCountryCode)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *accountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan accountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyAndRefresh(ctx, plan.Name.ValueString(), plan.AppleStoreCountryCode.ValueString(), &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op: SimpleMDM accounts cannot be deleted via the API.
// We simply remove the resource from Terraform state.
func (r *accountResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *accountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *accountResource) applyAndRefresh(ctx context.Context, name, countryCode string, plan *accountResourceModel, diags *diag.Diagnostics) {
	resp, err := simplemdmext.UpdateAccount(ctx, r.client, name, countryCode)
	if err != nil {
		diags.AddError("Error updating SimpleMDM Account", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d", resp.Data.ID))
	plan.Name = types.StringValue(resp.Data.Attributes.Name)
	plan.AppleStoreCountryCode = types.StringValue(resp.Data.Attributes.AppleStoreCountryCode)
}
