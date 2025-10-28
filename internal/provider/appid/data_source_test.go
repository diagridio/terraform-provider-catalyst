package appid_test

import (
	"fmt"
	"testing"

	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"go.uber.org/mock/gomock"
)

func TestMockAppIdDataSource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockResourceClientFactory(ctrl)),
				),
			},
			Steps: []resource.TestStep{
				// Create AppId resource
				{
					Config: testAccAppIdResourceConfig(projectName, appidName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_appid.test", "name", appidName),
						resource.TestCheckResourceAttr("catalyst_appid.test", "project_id", projectName),
					),
				},
				// Read AppId datasource
				{
					Config: testAccAppIdResourceConfig(projectName, appidName) +
						testAccAppIdDatasourceConfig(projectName, appidName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.catalyst_appid.test", "name", appidName),
						resource.TestCheckResourceAttr("data.catalyst_appid.test", "project_id", projectName),
					),
				},
			},
		})
}

func testAccAppIdDatasourceConfig(projectName, appidName string) string {
	return fmt.Sprintf(`
data "catalyst_appid" "test" {
  project_id = %q
  name       = %q
}
`, projectName, appidName)
}
