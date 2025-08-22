package organization_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	conductor_client "github.com/diagridio/diagrid-cloud-go/pkg/conductor/client"
	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
)

type productAttributes struct {
	Cra *conductor_client.ProductAttributes `json:"cra,omitempty"`
	Mcp *conductor_client.ProductAttributes `json:"mcp,omitempty"`
}

func TestMockOrganizationDataSource(t *testing.T) {
	var (
		orgID   = uuid.NewString()
		orgName = acctest.RandomWithPrefix("org")
		orgPlan = "cra:standard"
	)
	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(
						func(endpoint, apiKey string) (catalyst.Client, error) {
							ctrl := gomock.NewController(t)
							c := catalyst.NewMockClient(ctrl)

							attrs := &productAttributes{
								Cra: &conductor_client.ProductAttributes{
									Plan: lo.ToPtr(orgPlan),
								},
							}
							c.EXPECT().GetUserOrg(gomock.Any()).Return(
								&conductor_client.Organization{
									Data: conductor_client.OrganizationData{
										Id: lo.ToPtr(orgID),
										Attributes: &conductor_client.OrganizationAttributes{
											Name: lo.ToPtr(orgName),
											Products: (*struct {
												Cra *conductor_client.ProductAttributes `json:"cra,omitempty"`
												Mcp *conductor_client.ProductAttributes `json:"mcp,omitempty"`
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
						resource.TestCheckResourceAttr("data.catalyst_organization.test", "id", orgID),
						resource.TestCheckResourceAttr("data.catalyst_organization.test", "name", orgName),
						resource.TestCheckResourceAttr("data.catalyst_organization.test", "plan", orgPlan),
					),
				},
			},
		})
}

const testAccOrganizationDataSourceConfig = `
data "catalyst_organization" "test" {}
`
