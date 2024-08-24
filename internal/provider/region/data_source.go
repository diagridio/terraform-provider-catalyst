package region

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

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
			"id": schema.StringAttribute{
				MarkdownDescription: "Region identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Region name",
				Optional:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "Cloud provider",
				Optional:            true,
			},
			"cloud_provider_region": schema.StringAttribute{
				MarkdownDescription: "Cloud provider region",
				Optional:            true,
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

	tflog.Debug(ctx, "reading region data", map[string]interface{}{
		"id": model.GetID(),
	})

	regions, err := d.client.ListRegions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list regions", err.Error())
		return
	}

	tflog.Debug(ctx, "read regions data", map[string]interface{}{
		"regions": *regions,
	})

	for _, region := range *regions {
		if *region.Id == model.GetName() {
			// update the model
			model.SetID(*region.Id)
			model.SetName(*region.DisplayName)
			model.SetCloudProvider(*region.CloudProvider)
			model.SetCloudProviderRegion(*region.CloudProviderRegion)

			break
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
