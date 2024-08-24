package project

import (
	"context"
	"fmt"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cloudruntime_errors "github.com/diagridio/diagrid-cloud-go/cloudruntime/errors"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &projectDataSource{}

// projectDataSource defines the data source implementation.
type projectDataSource struct {
	client catalyst.Client
}

// projectDataSourceModel describes the data source data model.
type projectDataSourceModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Region         types.String `tfsdk:"region"`
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
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Organization id",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Optional:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Region",
				Optional:            true,
			},
			"managed_pubsub": schema.BoolAttribute{
				MarkdownDescription: "Managed pubsub component enabled",
				Optional:            true,
				Computed:            true,
			},
			"managed_kvstore": schema.BoolAttribute{
				MarkdownDescription: "Managed KV store component enabled",
				Optional:            true,
				Computed:            true,
			},
			"managed_workflow": schema.BoolAttribute{
				MarkdownDescription: "Managed workflow component enabled",
				Optional:            true,
				Computed:            true,
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
		if cloudruntime_errors.IsResourceNotFoundError(err) {
			tflog.Debug(ctx, "project not found", map[string]interface{}{
				"name": model.GetName(),
			})
			return
		}

		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("error reading project: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
