package provider

import (
	"context"
	"fmt"
	"time"

	simplemdm "github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &depServerSyncResource{}
	_ resource.ResourceWithConfigure = &depServerSyncResource{}
)

type depServerSyncResource struct {
	client *simplemdm.Client
}

type depServerSyncResourceModel struct {
	ID            types.String `tfsdk:"id"`
	DepServerID   types.String `tfsdk:"dep_server_id"`
	Triggers      types.Map    `tfsdk:"triggers"`
	LastTriggered types.String `tfsdk:"last_triggered"`
}

func DepServerSyncResource() resource.Resource {
	return &depServerSyncResource{}
}

func (r *depServerSyncResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dep_server_sync"
}

func (r *depServerSyncResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a sync of the given DEP server with Apple Business Manager. This is a fire-and-forget action: the sync runs once on Create. To re-trigger a sync, change the `triggers` map to taint the resource (e.g. via the `replace_triggered_by` lifecycle meta-argument).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Synthetic ID for the sync action.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dep_server_id": schema.StringAttribute{
				Required:    true,
				Description: "Required. ID of the DEP server to sync.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"triggers": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Optional arbitrary string map. Changing any value forces resource replacement, re-running the sync.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"last_triggered": schema.StringAttribute{
				Computed:    true,
				Description: "RFC3339 timestamp of the most recent sync invocation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *depServerSyncResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*simplemdm.Client)
}

func (r *depServerSyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan depServerSyncResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := simplemdmext.SyncDepServer(ctx, r.client, plan.DepServerID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to trigger DEP server sync", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("dep-sync-%s-%d", plan.DepServerID.ValueString(), time.Now().UnixNano()))
	plan.LastTriggered = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read is intentionally a no-op: the action does not have remote state.
func (r *depServerSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state depServerSyncResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update should not be reachable because all schema attributes are RequiresReplace.
func (r *depServerSyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan depServerSyncResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete simply removes the resource from state.
func (r *depServerSyncResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
