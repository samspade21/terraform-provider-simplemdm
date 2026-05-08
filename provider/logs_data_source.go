package provider

import (
	"context"
	"fmt"

	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdm"
	"github.com/DavidKrau/terraform-provider-simplemdm/internal/simplemdmext"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &logsDataSource{}
	_ datasource.DataSourceWithConfigure = &logsDataSource{}
)

type logsDataSource struct {
	client *simplemdm.Client
}

type logsDataSourceModel struct {
	Logs []logModel `tfsdk:"logs"`
}

type logModel struct {
	ID        types.String `tfsdk:"id"`
	Namespace types.String `tfsdk:"namespace"`
	Source    types.String `tfsdk:"source"`
	EventType types.String `tfsdk:"event_type"`
	Level     types.String `tfsdk:"level"`
	Message   types.String `tfsdk:"message"`
	At        types.String `tfsdk:"at"`
}

func LogsDataSource() datasource.DataSource {
	return &logsDataSource{}
}

func (d *logsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logs"
}

func (d *logsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches audit logs from your SimpleMDM account. Note: this data source fetches all logs and may be slow for accounts with large log volumes.",
		Blocks: map[string]schema.Block{
			"logs": schema.ListNestedBlock{
				Description: "Collection of log entries.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the log entry.",
						},
						"namespace": schema.StringAttribute{
							Computed:    true,
							Description: "The namespace of the log entry (e.g., 'mdm', 'user').",
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
							Description: "The log message.",
						},
						"at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when the log entry was created.",
						},
					},
				},
			},
		},
	}
}

func (d *logsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state logsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	logs, err := simplemdmext.ListLogs(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read SimpleMDM Logs",
			err.Error(),
		)
		return
	}

	state.Logs = make([]logModel, 0, len(logs))
	for _, l := range logs {
		state.Logs = append(state.Logs, logModel{
			ID:        types.StringValue(l.ID),
			Namespace: types.StringValue(l.Attributes.Namespace),
			Source:    types.StringValue(l.Attributes.Source),
			EventType: types.StringValue(l.Attributes.EventType),
			Level:     types.StringValue(l.Attributes.Level),
			Message:   types.StringValue(l.Attributes.Message),
			At:        types.StringValue(l.Attributes.At),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *logsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
