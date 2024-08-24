package organization_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"

	"github.com/diagridio/diagrid-cloud-go/pkg/client"
	diagrid_client "github.com/diagridio/diagrid-cloud-go/pkg/client"
	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
	"github.com/diagridio/terraform-provider-catalyst/internal/test/acceptance"
)

type productAttributes struct {
	Cra *diagrid_client.ProductAttributes `json:"cra,omitempty"`
	Mcp *client.ProductAttributes         `json:"mcp,omitempty"`
}

func TestAccOrganizationDataSource(t *testing.T) {
	t.Skip("skipping")
	resource.Test(t,
		resource.TestCase{
			PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Read testing
				{
					Config: testAccOrganizationDataSourceConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.catalyst_organization.test", "name", "terraform"),
						resource.TestCheckResourceAttr("data.catalyst_organization.test", "plan", "cra:standard"),
					),
				},
			},
		})
}

func TestMockOrganizationDataSource(t *testing.T) {
	resource.Test(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(
						func(endpoint, apiKey string) (catalyst.Client, error) {
							ctrl := gomock.NewController(t)
							c := catalyst.NewMockClient(ctrl)

							attrs := &productAttributes{
								Cra: &diagrid_client.ProductAttributes{
									Plan: lo.ToPtr("cra:standard"),
								},
							}
							c.EXPECT().GetUserOrg(gomock.Any()).Return(
								&diagrid_client.Organization{
									Data: diagrid_client.OrganizationData{
										Id: lo.ToPtr(uuid.NewString()),
										Attributes: &diagrid_client.OrganizationAttributes{
											Name: lo.ToPtr("terraform"),
											Products: (*struct {
												Cra *diagrid_client.ProductAttributes `json:"cra,omitempty"`
												Mcp *client.ProductAttributes         `json:"mcp,omitempty"`
											})(attrs),
										},
									},
								}, nil).
								AnyTimes()

							return c, nil
						}),
				),
			},
			Steps: []resource.TestStep{
				// Read testing
				{
					Config: testAccOrganizationDataSourceConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.catalyst_organization.test", "name", "terraform"),
						resource.TestCheckResourceAttr("data.catalyst_organization.test", "plan", "cra:standard"),
					),
				},
			},
		})
}

const testAccOrganizationDataSourceConfig = `
data "catalyst_organization" "test" {}
`
