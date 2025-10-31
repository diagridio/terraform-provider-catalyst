package kvstore

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
)

var _ datasource.DataSource = &kvstoreDataSource{}

type kvstoreDataSource struct {
	client catalyst.Client
}

func NewDataSource() datasource.DataSource {
	return &kvstoreDataSource{}
}

func (d *kvstoreDataSource) Metadata(ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_kvstore"
}

func (d *kvstoreDataSource) Schema(ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "KVStore data source",

		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "KVStore name",
				Required:            true,
			},
			"component_name": schema.StringAttribute{
				MarkdownDescription: "Component name",
				Computed:            true,
			},
			"create_component": schema.BoolAttribute{
				MarkdownDescription: "Create component",
				Computed:            true,
			},
			"scopes": schema.ListAttribute{
				MarkdownDescription: "Scopes",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status",
				Computed:            true,
			},
		},
	}
}

func (d *kvstoreDataSource) Configure(ctx context.Context,
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

func (d *kvstoreDataSource) Read(ctx context.Context,
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
			fmt.Sprintf("error reading kvstore datasource: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
