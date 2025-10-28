package component

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
)

var _ datasource.DataSource = &componentDataSource{}

type componentDataSource struct {
	client catalyst.Client
}

func NewDataSource() datasource.DataSource {
	return &componentDataSource{}
}

func (d *componentDataSource) Metadata(ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

func (d *componentDataSource) Schema(ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Component data source",

		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Component name",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Component type",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Component version",
				Computed:            true,
			},
			"spec": schema.StringAttribute{
				MarkdownDescription: "Component metadata as YAML string",
				Computed:            true,
			},
			"secret_store": schema.StringAttribute{
				MarkdownDescription: "Secret store for component authentication",
				Computed:            true,
			},
			"scopes": schema.ListAttribute{
				MarkdownDescription: "App IDs that can access this component",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the component",
				Computed:            true,
			},
		},
	}
}

func (d *componentDataSource) Configure(ctx context.Context,
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

func (d *componentDataSource) Read(ctx context.Context,
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
			fmt.Sprintf("error reading component datasource: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
