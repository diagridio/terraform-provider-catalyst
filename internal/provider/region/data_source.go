package region

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &dataSource{}

// dataSource defines the data source implementation.
type dataSource struct {
	client catalyst.Client
}

func NewDataSource() datasource.DataSource {
	return &dataSource{}
}

func (d *dataSource) Metadata(ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_region"
}

func (d *dataSource) Schema(ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Region data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Region name",
				Required:            true,
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Region host",
				Optional:            true,
			},
			"ingress": schema.StringAttribute{
				MarkdownDescription: "Region ingress",
				Optional:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Region location",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Region type",
				Optional:            true,
			},
			"join_token": schema.StringAttribute{
				MarkdownDescription: "Join token for the region",
				Computed:            true,
				Sensitive:           true,
			},
			"connected": schema.BoolAttribute{
				MarkdownDescription: "Whether the region is connected",
				Computed:            true,
			},
		},
	}
}

func (d *dataSource) Configure(ctx context.Context,
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
			fmt.Sprintf("Expected data.ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = providerData.Client
}

func (d *dataSource) Read(ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	model := NewModel()

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "reading region data",
		map[string]interface{}{
			"name": model.GetName(),
		})

	if err := read(ctx, d.client, model); err != nil {
		if diagrid_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "project not found", map[string]interface{}{
				"name": model.GetName(),
			})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading project datasource: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
