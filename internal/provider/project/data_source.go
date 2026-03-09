package project

import (
	"context"
	"fmt"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &projectDataSource{}

// projectDataSource defines the data source implementation.
type projectDataSource struct {
	client catalyst.Client
}

func NewDataSource() datasource.DataSource {
	return &projectDataSource{}
}

func (d *projectDataSource) Metadata(ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Project data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Optional:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Region",
				Optional:            true,
			},
			"grpc_endpoint": schema.StringAttribute{
				MarkdownDescription: "gRPC endpoint",
				Optional:            true,
				Computed:            true,
			},
			"http_endpoint": schema.StringAttribute{
				MarkdownDescription: "HTTP endpoint",
				Optional:            true,
				Computed:            true,
			},
			"default_agent_infrastructure_enabled": schema.BoolAttribute{
				MarkdownDescription: "Default agent infrastructure enabled",
				Computed:            true,
			},
			"default_kvstore_enabled": schema.BoolAttribute{
				MarkdownDescription: "Default KV store enabled",
				Computed:            true,
			},
			"default_pubsub_enabled": schema.BoolAttribute{
				MarkdownDescription: "Default pub/sub enabled",
				Computed:            true,
			},
			"default_workflow_store_enabled": schema.BoolAttribute{
				MarkdownDescription: "Default workflow store enabled",
				Computed:            true,
			},
			"disable_app_tunnels": schema.BoolAttribute{
				MarkdownDescription: "App tunnels disabled",
				Computed:            true,
			},
			"private_region": schema.BoolAttribute{
				MarkdownDescription: "Private region",
				Computed:            true,
			},
			"global_app_id_max_body_size": schema.StringAttribute{
				MarkdownDescription: "Maximum body size for HTTP and gRPC requests across all appids",
				Computed:            true,
			},
		},
	}
}

func (d *projectDataSource) Configure(ctx context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(data.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = providerData.Client
}

func (d *projectDataSource) Read(ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	model := NewModel()

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := read(ctx, d.client, model); err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading project datasource: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
