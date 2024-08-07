package organization

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/diagridio/diagrid-cloud-go/management"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &dataSource{}

// dataSource defines the data source implementation.
type dataSource struct {
	managementClient *management.ManagementClient
}

func NewDataSource() datasource.DataSource {
	return &dataSource{}
}

func (d *dataSource) Metadata(ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *dataSource) Schema(ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Organization data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Organization identifier",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Optional:            true,
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: "Organization plan",
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

	d.managementClient = providerData.ManagementClient
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

	// Read the user's current organization data
	org, err := d.managementClient.GetUserOrg(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read organization data", err.Error())
		return
	}

	tflog.Debug(ctx, "read organization data", map[string]interface{}{
		"id":   *org.Data.Id,
		"name": *org.Data.Attributes.Name,
	})

	// Save the organization data into the model
	model.SetID(*org.Data.Id)
	model.SetName(*org.Data.Attributes.Name)
	if org.Data.Attributes.Products.Cra != nil {
		model.SetPlan(*org.Data.Attributes.Products.Cra.Plan)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}
