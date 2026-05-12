package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type customDeclarationDataSource struct {
	client *simplemdm.Client
}

type customDeclarationDataSourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	DeclarationType        types.String `tfsdk:"declaration_type"`
	Payload                types.String `tfsdk:"payload"`
	UserScope              types.Bool   `tfsdk:"user_scope"`
	AttributeSupport       types.Bool   `tfsdk:"attribute_support"`
	EscapeAttributes       types.Bool   `tfsdk:"escape_attributes"`
	ActivationPredicate    types.String `tfsdk:"activation_predicate"`
	ReinstallAfterOsUpdate types.Bool   `tfsdk:"reinstall_after_os_update"`
	ProfileIdentifier      types.String `tfsdk:"profile_identifier"`
	GroupCount             types.Int64  `tfsdk:"group_count"`
	DeviceCount            types.Int64  `tfsdk:"device_count"`
}

var _ datasource.DataSource = &customDeclarationDataSource{}
var _ datasource.DataSourceWithConfigure = &customDeclarationDataSource{}

func CustomDeclarationDataSource() datasource.DataSource {
	return &customDeclarationDataSource{}
}

func (d *customDeclarationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_customdeclaration"
}

func (d *customDeclarationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Custom Declaration data source retrieves Declarative Device Management custom declarations from SimpleMDM.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the declaration to retrieve.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "A name for the custom declaration.",
			},
			"declaration_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of declaration being defined.",
			},
			"payload": schema.StringAttribute{
				Computed:    true,
				Description: "The JSON payload for the declaration.",
			},
			"user_scope": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the declaration is scoped to users (true) or devices (false).",
			},
			"attribute_support": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether variable expansion is enabled for the declaration payload.",
			},
			"escape_attributes": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether custom variable values are escaped before being delivered.",
			},
			"activation_predicate": schema.StringAttribute{
				Computed:    true,
				Description: "Predicate that controls when the declaration activates on a device.",
			},
			"reinstall_after_os_update": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to reinstall the declaration after macOS updates.",
			},
			"profile_identifier": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier assigned by SimpleMDM for tracking the declaration profile.",
			},
			"group_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of device groups currently assigned to the declaration.",
			},
			"device_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of devices currently assigned to the declaration.",
			},
		},
	}
}

func (d *customDeclarationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*simplemdm.Client)
}

func (d *customDeclarationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customDeclarationDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	declarationID := state.ID.ValueString()

	// SimpleMDM does not expose GET /custom_declarations/{id}; list+filter is
	// the only way to read a single declaration's metadata.
	item, err := findCustomDeclarationByID(ctx, d.client, declarationID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SimpleMDM custom declaration", err.Error())
		return
	}
	if item == nil {
		resp.Diagnostics.AddError("Custom declaration not found", fmt.Sprintf("No custom declaration with id %q found in SimpleMDM", declarationID))
		return
	}

	// /download gives us declaration_type (top-level Type) and the assembled
	// payload (with SimpleMDM-injected declaration_name / activation_predicate
	// keys stripped by parseDeclarationEnvelope).
	rawEnvelope, err := downloadCustomDeclarationPayload(ctx, d.client, declarationID)
	if err != nil {
		resp.Diagnostics.AddError("Error downloading SimpleMDM custom declaration payload", err.Error())
		return
	}
	declType, activationPredicate, cleanedPayload := parseDeclarationEnvelope(rawEnvelope)

	declaration := customDeclarationResponse{}
	declaration.Data.ID = json.Number(declarationID)
	declaration.Data.Attributes = item.Attributes
	if declType != "" {
		declaration.Data.Attributes.DeclarationType = declType
	}
	if activationPredicate != "" && declaration.Data.Attributes.ActivationPredicate == "" {
		declaration.Data.Attributes.ActivationPredicate = activationPredicate
	}
	if len(cleanedPayload) > 0 {
		declaration.Data.Attributes.Payload = cleanedPayload
	}

	var model customDeclarationResourceModel
	if diags := model.refreshFromResponse(ctx, &declaration); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Copy data from resource model into data source state.
	state.ID = model.ID
	state.Name = model.Name
	state.DeclarationType = model.DeclarationType
	state.Payload = model.Payload
	state.UserScope = model.UserScope
	state.AttributeSupport = model.AttributeSupport
	state.EscapeAttributes = model.EscapeAttributes
	state.ActivationPredicate = model.ActivationPredicate
	state.ReinstallAfterOsUpdate = model.ReinstallAfterOsUpdate
	state.ProfileIdentifier = model.ProfileIdentifier
	state.GroupCount = model.GroupCount
	state.DeviceCount = model.DeviceCount

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
