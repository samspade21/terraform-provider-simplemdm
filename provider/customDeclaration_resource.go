package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type customDeclarationResource struct {
	client *simplemdm.Client
}

type customDeclarationResourceModel struct {
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

type customDeclarationAttributes struct {
	Name                   string          `json:"name"`
	DeclarationType        string          `json:"declaration_type"`
	Payload                json.RawMessage `json:"payload"`
	UserScope              *bool           `json:"user_scope"`
	AttributeSupport       *bool           `json:"attribute_support"`
	EscapeAttributes       *bool           `json:"escape_attributes"`
	ActivationPredicate    string          `json:"activation_predicate"`
	ReinstallAfterOsUpdate *bool           `json:"reinstall_after_os_update"`
	ProfileIdentifier      string          `json:"profile_identifier"`
	GroupCount             *int64          `json:"group_count"`
	DeviceCount            *int64          `json:"device_count"`
}

type customDeclarationResponse struct {
	Data struct {
		ID         json.Number                 `json:"id"`
		Attributes customDeclarationAttributes `json:"attributes"`
	} `json:"data"`
}

// customDeclarationPayload represents the multipart form data for Create/Update
type customDeclarationPayload struct {
	Name                   string `json:"name"`
	DeclarationType        string `json:"declaration_type"`
	Payload                []byte `json:"payload"`
	UserScope              *bool  `json:"user_scope,omitempty"`
	AttributeSupport       *bool  `json:"attribute_support,omitempty"`
	EscapeAttributes       *bool  `json:"escape_attributes,omitempty"`
	ActivationPredicate    string `json:"activation_predicate,omitempty"`
	ReinstallAfterOsUpdate *bool  `json:"reinstall_after_os_update,omitempty"`
}

func CustomDeclarationResource() resource.Resource {
	return &customDeclarationResource{}
}

var _ resource.Resource = &customDeclarationResource{}
var _ resource.ResourceWithConfigure = &customDeclarationResource{}
var _ resource.ResourceWithImportState = &customDeclarationResource{}

func (r *customDeclarationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_customdeclaration"
}

func (r *customDeclarationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Custom Declaration resource manages Declarative Device Management custom declarations in SimpleMDM.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A name for the custom declaration.",
			},
			"declaration_type": schema.StringAttribute{
				Required:    true,
				Description: "The type of declaration being defined (e.g., com.apple.configuration.management.status-subscriptions).",
			},
			"payload": schema.StringAttribute{
				Required:    true,
				Description: "The JSON payload for the declaration. Stored verbatim; semantically-equal JSON with different whitespace or key order will register as drift.",
			},
			"user_scope": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the declaration is scoped to users (true) or devices (false). Defaults to true.",
				Default:     booldefault.StaticBool(true),
			},
			"attribute_support": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable variable expansion when processing the declaration payload. Defaults to false.",
				Default:     booldefault.StaticBool(false),
			},
			"escape_attributes": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Escape the values of custom variables within the payload before delivery. Defaults to false.",
				Default:     booldefault.StaticBool(false),
			},
			"activation_predicate": schema.StringAttribute{
				Optional:    true,
				Description: "Predicate format string that controls when the declaration activates on a device.",
			},
			"reinstall_after_os_update": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to reinstall the declaration after macOS updates. Defaults to false.",
				Default:     booldefault.StaticBool(false),
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

func (r *customDeclarationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*simplemdm.Client)
}

func (r *customDeclarationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customDeclarationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := buildCustomDeclarationPayload(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add required fields
	if err := writer.WriteField("name", payload.Name); err != nil {
		resp.Diagnostics.AddError("Error building multipart request", err.Error())
		return
	}

	if err := writer.WriteField("declaration_type", payload.DeclarationType); err != nil {
		resp.Diagnostics.AddError("Error building multipart request", err.Error())
		return
	}

	// Add payload as a file part
	if len(payload.Payload) > 0 {
		part, err := writer.CreateFormFile("payload", "declaration.json")
		if err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
		if _, err := part.Write(payload.Payload); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	// Add optional fields
	if payload.UserScope != nil {
		if err := writer.WriteField("user_scope", fmt.Sprintf("%t", *payload.UserScope)); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if payload.AttributeSupport != nil {
		if err := writer.WriteField("attribute_support", fmt.Sprintf("%t", *payload.AttributeSupport)); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if payload.EscapeAttributes != nil {
		if err := writer.WriteField("escape_attributes", fmt.Sprintf("%t", *payload.EscapeAttributes)); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if payload.ActivationPredicate != "" {
		if err := writer.WriteField("activation_predicate", payload.ActivationPredicate); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if payload.ReinstallAfterOsUpdate != nil {
		if err := writer.WriteField("reinstall_after_os_update", fmt.Sprintf("%t", *payload.ReinstallAfterOsUpdate)); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if err := writer.Close(); err != nil {
		resp.Diagnostics.AddError("Error building multipart request", err.Error())
		return
	}

	url := fmt.Sprintf("https://%s/api/v1/custom_declarations", r.client.HostName)
	httpReq, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SimpleMDM custom declaration request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// SimpleMDM returns 201 Created. The previous 200-then-fallback-to-201
	// pattern re-posted the request with an already-consumed body buffer,
	// which caused a 400 response after the original POST had already
	// successfully created a record server-side (orphans on every apply).
	responseBody, err := r.client.RequestResponse201(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SimpleMDM custom declaration", err.Error())
		return
	}

	var declaration customDeclarationResponse
	if err := json.Unmarshal(responseBody, &declaration); err != nil {
		resp.Diagnostics.AddError("Error parsing SimpleMDM custom declaration response", err.Error())
		return
	}

	declarationID := declaration.Data.ID.String()
	if declarationID == "" {
		resp.Diagnostics.AddError(
			"Error creating SimpleMDM custom declaration",
			"Response from SimpleMDM did not include an id field",
		)
		return
	}

	// SimpleMDM's create response omits payload, declaration_type, and
	// activation_predicate — those come from the plan. refreshFromResponse
	// preserves the plan's values for any field the API omits. The user-
	// supplied payload string is preserved verbatim — normalising it would
	// fail "Provider produced inconsistent result after apply" because
	// Required attributes can't be modified vs the configured value.
	if diags := plan.refreshFromResponse(ctx, &declaration); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *customDeclarationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customDeclarationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	declarationID := state.ID.ValueString()

	// SimpleMDM does not expose GET /custom_declarations/{id}; list+filter is
	// the only way to read a single declaration's metadata.
	item, err := findCustomDeclarationByID(ctx, r.client, declarationID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SimpleMDM custom declaration", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// /download gives us Type (declaration_type) and the assembled payload.
	rawEnvelope, err := downloadCustomDeclarationPayload(ctx, r.client, declarationID)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error downloading SimpleMDM custom declaration payload", err.Error())
		return
	}

	declType, _, _ := parseDeclarationEnvelope(rawEnvelope)

	declaration := customDeclarationResponse{}
	declaration.Data.ID = json.Number(declarationID)
	declaration.Data.Attributes = item.Attributes
	if declType != "" {
		declaration.Data.Attributes.DeclarationType = declType
	}
	// Don't pass attributes.Payload — refreshFromResponse would normalise it
	// and that fights the user's verbatim-store schema. We preserve the
	// previous state value below.

	// refreshFromResponse leaves Payload alone when attributes.Payload is
	// empty, so the verbatim value from prior state is preserved.
	if diags := state.refreshFromResponse(ctx, &declaration); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *customDeclarationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customDeclarationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := buildCustomDeclarationPayload(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add fields to update
	if err := writer.WriteField("name", payload.Name); err != nil {
		resp.Diagnostics.AddError("Error building multipart request", err.Error())
		return
	}

	if err := writer.WriteField("declaration_type", payload.DeclarationType); err != nil {
		resp.Diagnostics.AddError("Error building multipart request", err.Error())
		return
	}

	// Add payload as a file part if present
	if len(payload.Payload) > 0 {
		part, err := writer.CreateFormFile("payload", "declaration.json")
		if err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
		if _, err := part.Write(payload.Payload); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	// Add optional fields
	if payload.UserScope != nil {
		if err := writer.WriteField("user_scope", fmt.Sprintf("%t", *payload.UserScope)); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if payload.AttributeSupport != nil {
		if err := writer.WriteField("attribute_support", fmt.Sprintf("%t", *payload.AttributeSupport)); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if payload.EscapeAttributes != nil {
		if err := writer.WriteField("escape_attributes", fmt.Sprintf("%t", *payload.EscapeAttributes)); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if payload.ActivationPredicate != "" {
		if err := writer.WriteField("activation_predicate", payload.ActivationPredicate); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if payload.ReinstallAfterOsUpdate != nil {
		if err := writer.WriteField("reinstall_after_os_update", fmt.Sprintf("%t", *payload.ReinstallAfterOsUpdate)); err != nil {
			resp.Diagnostics.AddError("Error building multipart request", err.Error())
			return
		}
	}

	if err := writer.Close(); err != nil {
		resp.Diagnostics.AddError("Error building multipart request", err.Error())
		return
	}

	url := fmt.Sprintf("https://%s/api/v1/custom_declarations/%s", r.client.HostName, plan.ID.ValueString())
	httpReq, err := http.NewRequest(http.MethodPatch, url, &body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SimpleMDM custom declaration request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	responseBody, err := r.client.RequestResponse200(httpReq)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error updating SimpleMDM custom declaration", err.Error())
		return
	}

	var declaration customDeclarationResponse
	if err := json.Unmarshal(responseBody, &declaration); err != nil {
		resp.Diagnostics.AddError("Error parsing SimpleMDM custom declaration response", err.Error())
		return
	}

	// The PATCH response is the list-shape record (no payload, no
	// declaration_type, no activation_predicate); refreshFromResponse
	// preserves the planned values for fields the API omits.
	if diags := plan.refreshFromResponse(ctx, &declaration); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *customDeclarationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customDeclarationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://%s/api/v1/custom_declarations/%s", r.client.HostName, state.ID.ValueString())
	httpReq, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SimpleMDM custom declaration request", err.Error())
		return
	}

	_, err = r.client.RequestResponse204(httpReq)
	if err != nil {
		if isNotFoundError(err) {
			return
		}

		resp.Diagnostics.AddError("Error deleting SimpleMDM custom declaration", err.Error())
	}
}

func (r *customDeclarationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildCustomDeclarationPayload(ctx context.Context, model *customDeclarationResourceModel) (*customDeclarationPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Validate required fields (BUG-CD-007)
	if model.Name.IsNull() || model.Name.ValueString() == "" {
		diags.AddError("Missing required field", "name is required")
	}
	if model.DeclarationType.IsNull() || model.DeclarationType.ValueString() == "" {
		diags.AddError("Missing required field", "declaration_type is required")
	}
	if model.Payload.IsNull() || model.Payload.ValueString() == "" {
		diags.AddError("Missing required field", "payload is required")
	}

	if diags.HasError() {
		return nil, diags
	}

	normalizedPayload, err := normalizeJSON(model.Payload.ValueString(), "payload", model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid JSON payload", fmt.Sprintf("Unable to parse declaration payload: %s", err))
		return nil, diags
	}

	payload := &customDeclarationPayload{
		Name:            model.Name.ValueString(),
		DeclarationType: model.DeclarationType.ValueString(),
		Payload:         []byte(normalizedPayload),
	}

	if !model.UserScope.IsNull() {
		userScope := model.UserScope.ValueBool()
		payload.UserScope = &userScope
	}

	if !model.AttributeSupport.IsNull() {
		attributeSupport := model.AttributeSupport.ValueBool()
		payload.AttributeSupport = &attributeSupport
	}

	if !model.EscapeAttributes.IsNull() {
		escapeAttributes := model.EscapeAttributes.ValueBool()
		payload.EscapeAttributes = &escapeAttributes
	}

	if !model.ActivationPredicate.IsNull() && model.ActivationPredicate.ValueString() != "" {
		payload.ActivationPredicate = model.ActivationPredicate.ValueString()
	}

	if !model.ReinstallAfterOsUpdate.IsNull() {
		reinstall := model.ReinstallAfterOsUpdate.ValueBool()
		payload.ReinstallAfterOsUpdate = &reinstall
	}

	return payload, diags
}

// refreshFromResponse maps API attributes onto the resource model. Fields the
// API never returns (declaration_type, activation_predicate, payload) are left
// alone when the response is empty — callers are responsible for populating
// them from the plan/state before persisting.
func (m *customDeclarationResourceModel) refreshFromResponse(ctx context.Context, response *customDeclarationResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	attributes := response.Data.Attributes

	m.ID = types.StringValue(response.Data.ID.String())
	if attributes.Name != "" {
		m.Name = types.StringValue(attributes.Name)
	}
	if attributes.DeclarationType != "" {
		m.DeclarationType = types.StringValue(attributes.DeclarationType)
	}

	if attributes.UserScope != nil {
		m.UserScope = types.BoolValue(*attributes.UserScope)
	}

	if attributes.AttributeSupport != nil {
		m.AttributeSupport = types.BoolValue(*attributes.AttributeSupport)
	}

	if attributes.EscapeAttributes != nil {
		m.EscapeAttributes = types.BoolValue(*attributes.EscapeAttributes)
	}

	if attributes.ActivationPredicate != "" {
		m.ActivationPredicate = types.StringValue(attributes.ActivationPredicate)
	}

	if attributes.ReinstallAfterOsUpdate != nil {
		m.ReinstallAfterOsUpdate = types.BoolValue(*attributes.ReinstallAfterOsUpdate)
	}

	if attributes.ProfileIdentifier != "" {
		m.ProfileIdentifier = types.StringValue(attributes.ProfileIdentifier)
	} else {
		m.ProfileIdentifier = types.StringNull()
	}

	if attributes.GroupCount != nil {
		m.GroupCount = types.Int64Value(*attributes.GroupCount)
	} else {
		m.GroupCount = types.Int64Null()
	}

	if attributes.DeviceCount != nil {
		m.DeviceCount = types.Int64Value(*attributes.DeviceCount)
	} else {
		m.DeviceCount = types.Int64Null()
	}

	if len(attributes.Payload) > 0 {
		m.Payload = types.StringValue(string(attributes.Payload))
	}

	return diags
}

func downloadCustomDeclarationPayload(ctx context.Context, client *simplemdm.Client, declarationID string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://%s/api/v1/custom_declarations/%s/download", client.HostName, declarationID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := client.RequestResponse200(httpReq)
	if err != nil {
		return nil, err
	}

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil, nil
	}

	return json.RawMessage(trimmed), nil
}

// parseDeclarationEnvelope extracts the user-meaningful fields from a
// /custom_declarations/{id}/download response, which is the assembled DDM
// envelope:
//
//	{"Type":"…","Identifier":"…","ServerToken":"…","Payload":{<user fields>+
//	 "declaration_name":"…","activation_predicate":null|"…"}}
//
// Returns (declaration_type, activation_predicate, raw user payload with the
// SimpleMDM-injected declaration_name and activation_predicate keys stripped).
// All three return values are best-effort — callers should treat empty values
// as "unknown" and fall back to plan / state.
func parseDeclarationEnvelope(raw json.RawMessage) (string, string, json.RawMessage) {
	if len(raw) == 0 {
		return "", "", nil
	}

	var envelope struct {
		Type    string          `json:"Type"`
		Payload json.RawMessage `json:"Payload"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return "", "", nil
	}

	if len(envelope.Payload) == 0 {
		return envelope.Type, "", nil
	}

	decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
	decoder.UseNumber()
	var asMap map[string]json.RawMessage
	if err := decoder.Decode(&asMap); err != nil {
		return envelope.Type, "", envelope.Payload
	}

	var activationPredicate string
	if v, ok := asMap["activation_predicate"]; ok {
		// Pull the value out, then strip; ignore nulls.
		if !bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			_ = json.Unmarshal(v, &activationPredicate)
		}
		delete(asMap, "activation_predicate")
	}
	delete(asMap, "declaration_name")

	cleaned, err := json.Marshal(asMap)
	if err != nil {
		return envelope.Type, activationPredicate, envelope.Payload
	}
	return envelope.Type, activationPredicate, cleaned
}

// findCustomDeclarationByID pages through /api/v1/custom_declarations looking
// for the given id. Returns (nil, nil) when the declaration cannot be found —
// callers should treat that as "resource gone, drop from state".
//
// SimpleMDM does not support GET /custom_declarations/{id}; the only way to
// read a single declaration's attributes is to filter the list.
func findCustomDeclarationByID(ctx context.Context, client *simplemdm.Client, declarationID string) (*customDeclarationDataList, error) {
	startingAfter := ""
	limit := 100
	for {
		url := fmt.Sprintf("https://%s/api/v1/custom_declarations?limit=%d", client.HostName, limit)
		if startingAfter != "" {
			url += fmt.Sprintf("&starting_after=%s", startingAfter)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		body, err := client.RequestResponse200(req)
		if err != nil {
			return nil, err
		}
		page, hasMore, err := simplemdm.DecodeList[customDeclarationDataList](body)
		if err != nil {
			return nil, err
		}
		for i := range page {
			if page[i].idString() == declarationID {
				return &page[i], nil
			}
		}
		if !hasMore || len(page) == 0 {
			return nil, nil
		}
		startingAfter = page[len(page)-1].idString()
	}
}

func normalizeJSON(input string, fieldName string, declarationID string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		contextInfo := ""
		if declarationID != "" {
			contextInfo = fmt.Sprintf(" in declaration %s", declarationID)
		}
		return "", fmt.Errorf("field '%s'%s: expected JSON object or array", fieldName, contextInfo)
	}

	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		contextInfo := ""
		if declarationID != "" {
			contextInfo = fmt.Sprintf(" in declaration %s", declarationID)
		}
		return "", fmt.Errorf("field '%s'%s: %w", fieldName, contextInfo, err)
	}

	normalized, err := json.Marshal(value)
	if err != nil {
		contextInfo := ""
		if declarationID != "" {
			contextInfo = fmt.Sprintf(" in declaration %s", declarationID)
		}
		return "", fmt.Errorf("field '%s'%s: %w", fieldName, contextInfo, err)
	}

	return string(normalized), nil
}
