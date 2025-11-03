package component_test

import (
	"fmt"
	"testing"

	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"go.uber.org/mock/gomock"
)

func TestMockComponentDataSource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockResourceClientFactory(ctrl)),
				),
			},
			Steps: []resource.TestStep{
				// Create Component resource
				{
					Config: testAccComponentResourceConfig(projectName, componentName, componentType, componentVersion),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_component.test", "name", componentName),
						resource.TestCheckResourceAttr("catalyst_component.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.type", componentType),
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.version", componentVersion),
					),
				},
				// Read Component datasource
				{
					Config: testAccComponentResourceConfig(projectName, componentName, componentType, componentVersion) +
						testAccComponentDatasourceConfig(projectName, componentName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.catalyst_component.test", "name", componentName),
						resource.TestCheckResourceAttr("data.catalyst_component.test", "project_name", projectName),
						resource.TestCheckResourceAttr("data.catalyst_component.test", "spec.type", componentType),
						resource.TestCheckResourceAttr("data.catalyst_component.test", "spec.version", componentVersion),
					),
				},
			},
		})
}

func testAccComponentDatasourceConfig(projectName, componentName string) string {
	return fmt.Sprintf(`
data "catalyst_component" "test" {
  project_name = %q
  name         = %q
}
`, projectName, componentName)
}
