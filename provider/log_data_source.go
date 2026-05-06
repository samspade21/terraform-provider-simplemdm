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
	_ datasource.DataSource              = &logDataSource{}
	_ datasource.DataSourceWithConfigure = &logDataSource{}
)

type logDataSource struct {
	client *simplemdm.Client
}

type logDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Namespace types.String `tfsdk:"namespace"`
	Source    types.String `tfsdk:"source"`
	EventType types.String `tfsdk:"event_type"`
	Level     types.String `tfsdk:"level"`
	Message   types.String `tfsdk:"message"`
	At        types.String `tfsdk:"at"`
	Metadata  types.String `tfsdk:"metadata"`
}

func LogDataSource() datasource.DataSource {
	return &logDataSource{}
}

func (d *logDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log"
}

func (d *logDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a specific SimpleMDM log entry by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Required. The ID of the log entry to retrieve.",
			},
			"namespace": schema.StringAttribute{
				Computed:    true,
				Description: "The namespace of the log entry (e.g. 'admin', 'device').",
			},
			"source": schema.StringAttribute{
				Computed:    true,
				Description: "The source of the log entry.",
			},
			"event_type": schema.StringAttribute{
				Computed:    true,
				Description: "The event type of the log entry (e.g. 'user.signed_in', 'script.ran').",
			},
			"level": schema.StringAttribute{
				Computed:    true,
				Description: "The severity level of the log entry.",
			},
			"message": schema.StringAttribute{
				Computed:    true,
				Description: "The log message body.",
			},
			"at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the log entry was created.",
			},
			"metadata": schema.StringAttribute{
				Computed:    true,
				Description: "Raw JSON metadata associated with the log entry.",
			},
		},
	}
}

func (d *logDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state logDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	logResp, err := simplemdmext.GetLog(ctx, d.client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read SimpleMDM log entry", err.Error())
		return
	}

	state.ID = types.StringValue(logResp.Data.ID)
	state.Namespace = types.StringValue(logResp.Data.Attributes.Namespace)
	state.Source = types.StringValue(logResp.Data.Attributes.Source)
	state.EventType = types.StringValue(logResp.Data.Attributes.EventType)
	state.Level = types.StringValue(logResp.Data.Attributes.Level)
	state.Message = types.StringValue(logResp.Data.Attributes.Message)
	state.At = types.StringValue(logResp.Data.Attributes.At)
	if len(logResp.Data.Attributes.Metadata) > 0 {
		state.Metadata = types.StringValue(string(logResp.Data.Attributes.Metadata))
	} else {
		state.Metadata = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *logDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*simplemdm.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *simplemdm.Client, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}
