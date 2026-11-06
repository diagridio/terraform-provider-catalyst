package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/diagridio/terraform-provider-diagrid/internal/diagrid"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/appid"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/component"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/configuration"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/data"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/kvstore"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/organization"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/project"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/pubsub"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/region"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/resiliency"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/serviceaccount"
	"github.com/diagridio/terraform-provider-diagrid/internal/provider/subscription"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &diagridProvider{}
var _ provider.ProviderWithFunctions = &diagridProvider{}

var (
	// ProdAPIEndpoint is the Base URL for Diagrid Production API endpoint.
	ProdAPIEndpoint = "https://api.r1.diagrid.io"
)

const (
	// ProviderName is the name of the provider.
	ProviderName = "diagrid"
)

type ClientFactory func(endpoint, apiKey string, tlsSkipVerify bool) (diagrid.Client, error)

type Provider interface {
	provider.Provider

	WithClientFactory(ClientFactory) Provider
}

// diagridProvider defines the provider implementation.
type diagridProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version       string
	clientFactory ClientFactory
}

// diagridModel describes the provider data model.
type diagridModel struct {
	Endpoint      types.String `tfsdk:"endpoint"`
	APIKey        types.String `tfsdk:"api_key"`
	TLSSkipVerify types.Bool   `tfsdk:"tls_skip_verify"`
}

func New(version string) Provider {
	return &diagridProvider{
		version:       version,
		clientFactory: diagrid.NewClient,
	}
}

func (p *diagridProvider) WithClientFactory(f ClientFactory) Provider {
	p.clientFactory = f
	return p
}

func (p *diagridProvider) Metadata(ctx context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse) {
	resp.TypeName = ProviderName
	resp.Version = p.version

	tflog.Debug(ctx, "Metadata response", map[string]interface{}{
		"response": *resp,
	})
}

func (p *diagridProvider) Schema(ctx context.Context,
	req provider.SchemaRequest,
	resp *provider.SchemaResponse) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				//Required:            true,
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "This is the Diagrid API key. Alternatively, this can also be specified using the `DIAGRID_API_KEY` environment variable.",
			},
			"endpoint": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Endpoint is the URL of Diagrid. Alternatively, this can also be specified using the `DIAGRID_API_ENDPOINT` environment variable.",
			},
			"tls_skip_verify": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Skip TLS certificate verification. This is useful for self-signed certificates. Alternatively, this can also be specified using the `DIAGRID_TLS_SKIP_VERIFY` environment variable.",
			},
		},
	}

	tflog.Debug(ctx, "Schema response", map[string]interface{}{
		"response": *resp,
	})
}

func (p *diagridProvider) Configure(ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse) {
	// Check environment variables
	apiKey := os.Getenv("DIAGRID_API_KEY")
	endpoint, ok := os.LookupEnv("DIAGRID_API_ENDPOINT")
	if !ok {
		// default to prod endpoint
		endpoint = ProdAPIEndpoint
	}

	tlsSkipVerify := false
	if v := os.Getenv("DIAGRID_TLS_SKIP_VERIFY"); v == "true" || v == "1" {
		tlsSkipVerify = true
	}

	var model diagridModel

	// Read the provider configuration from the request.
	diagnostics := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diagnostics...)
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
	if !model.TLSSkipVerify.IsNull() {
		tlsSkipVerify = model.TLSSkipVerify.ValueBool()
	}

	c, err := p.clientFactory(endpoint, apiKey, tlsSkipVerify)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Diagrid client",
			err.Error(),
		)
		return
	}

	providerData := data.ProviderData{
		Client: c,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *diagridProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		appid.NewResource,
		component.NewResource,
		configuration.NewResource,
		kvstore.NewResource,
		project.NewResource,
		pubsub.NewResource,
		region.NewResource,
		resiliency.NewResource,
		serviceaccount.NewResource,
		subscription.NewResource,
	}
}

func (p *diagridProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		appid.NewDataSource,
		component.NewDataSource,
		configuration.NewDataSource,
		kvstore.NewDataSource,
		organization.NewDataSource,
		project.NewDataSource,
		pubsub.NewDataSource,
		region.NewDataSource,
		resiliency.NewDataSource,
		serviceaccount.NewDataSource,
		subscription.NewDataSource,
	}
}

func (p *diagridProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}
