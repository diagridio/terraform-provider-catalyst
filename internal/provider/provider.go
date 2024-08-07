package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/diagridio/diagrid-cloud-go/cloudruntime"
	"github.com/diagridio/diagrid-cloud-go/management"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/data"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/organization"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/project"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/region"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &catalystProvider{}
var _ provider.ProviderWithFunctions = &catalystProvider{}

var (
	// ProdAPIEndpoint is the Base URL for Catalyst Production API endpoint
	ProdAPIEndpoint = "https://api.diagrid.io"
)

const (
	// ProviderName is the name of the provider.
	ProviderName = "catalyst"
)

// catalyst defines the provider implementation.
type catalystProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// catalystModel describes the provider data model.
type catalystModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &catalystProvider{
			version: version,
		}
	}
}
func (p *catalystProvider) Metadata(ctx context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse) {
	resp.TypeName = ProviderName
	resp.Version = p.version

	tflog.Debug(ctx, "Metadata response", map[string]interface{}{
		"response": *resp,
	})
}

func (p *catalystProvider) Schema(ctx context.Context,
	req provider.SchemaRequest,
	resp *provider.SchemaResponse) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				//Required:            true,
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "This is the Catalyst API key. Alternatively, this can also be specified using the `CATALYST_API_KEY` environment variable.",
			},
			"endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Endpoint is the URL of Catalyst. Alternatively, this can also be specified using the `CATALYST_API_ENDPOINT` environment variable.",
			},
		},
	}

	tflog.Debug(ctx, "Schema response", map[string]interface{}{
		"response": *resp,
	})
}

func (p *catalystProvider) Configure(ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse) {
	// Check environment variables
	apiKey := os.Getenv("CATALYST_API_KEY")
	endpoint, ok := os.LookupEnv("CATALYST_API_ENDPOINT")
	if !ok {
		// default to prod endpoint
		endpoint = ProdAPIEndpoint
	}

	var model catalystModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Error reading provider configuration")
		return
	}

	// Check configuration data, which should take precedence over
	// environment variable data, if found.
	if model.APIKey.ValueString() != "" {
		apiKey = model.APIKey.ValueString()
	}
	if model.Endpoint.ValueString() != "" {
		endpoint = model.Endpoint.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key Configuration",
			"While configuring the provider, the API key was not found in "+
				"the CATALYST_API_KEY environment variable or provider "+
				"configuration block api_key attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Endpoint Configuration",
			"While configuring the provider, the endpoint was not found in "+
				"the CATALYST_API_ENDPOINT environment variable or provider "+
				"configuration block endpoint attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	// Example client configuration for data sources and resources
	maxRetries := 1
	mc, err := management.NewManagementClientWithExponentialBackoff(http.DefaultClient,
		endpoint,
		maxRetries,
		management.WithAPIKeyToken(apiKey))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating management client",
			err.Error(),
		)
		return
	}

	catalystClient, err := cloudruntime.NewCloudruntimeClientWithExponentialBackoff(http.DefaultClient,
		endpoint,
		maxRetries,
		cloudruntime.WithAPIKeyToken(apiKey))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating catalyst client",
			err.Error(),
		)
		return
	}

	providerData := data.ProviderData{
		ManagementClient: mc,
		CatalystClient:   catalystClient,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *catalystProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		project.NewResource,
	}
}

func (p *catalystProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		organization.NewDataSource,
		project.NewDataSource,
		region.NewDataSource,
	}
}

func (p *catalystProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}
