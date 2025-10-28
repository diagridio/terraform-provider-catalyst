package appid

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
)

var _ datasource.DataSource = &appidDataSource{}

type appidDataSource struct {
	client catalyst.Client
}

func NewDataSource() datasource.DataSource {
	return &appidDataSource{}
}

func (d *appidDataSource) Metadata(ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_appid"
}

func (d *appidDataSource) Schema(ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "AppID data source",

		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "AppID name",
				Required:            true,
			},
			"app_config": schema.StringAttribute{
				MarkdownDescription: "App configuration name",
				Computed:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol (e.g., 'http', 'grpc')",
				Computed:            true,
			},
			"api_token_revision": schema.Int64Attribute{
				MarkdownDescription: "API token revision",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the AppID",
				Computed:            true,
			},
			"app_endpoint": schema.SingleNestedAttribute{
				MarkdownDescription: "Application endpoint configuration",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: "Application endpoint URL",
						Computed:            true,
					},
					"token": schema.StringAttribute{
						MarkdownDescription: "Authentication token",
						Computed:            true,
						Sensitive:           true,
					},
					"token_header": schema.StringAttribute{
						MarkdownDescription: "Header name for the authentication token",
						Computed:            true,
					},
					"client_timeout_seconds": schema.Int64Attribute{
						MarkdownDescription: "Client timeout in seconds",
						Computed:            true,
					},
				},
			},
			"health_check": schema.SingleNestedAttribute{
				MarkdownDescription: "Health check configuration",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						MarkdownDescription: "Health check path",
						Computed:            true,
					},
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Whether the probe is enabled",
						Computed:            true,
					},
					"failure_threshold": schema.Int64Attribute{
						MarkdownDescription: "Number of failures before marking as unhealthy",
						Computed:            true,
					},
					"interval_seconds": schema.Int64Attribute{
						MarkdownDescription: "Interval between probe checks in seconds",
						Computed:            true,
					},
					"timeout_ms": schema.Int64Attribute{
						MarkdownDescription: "Timeout for each probe check in milliseconds",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *appidDataSource) Configure(ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(data.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected data.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = providerData.Client
}

func (d *appidDataSource) Read(ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	model := NewModel()

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := read(ctx, d.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading appid datasource: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
