package organization_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/diagridio/terraform-provider-catalyst/internal/test/acceptance"
)

func TestAccOrganizationDataSource(t *testing.T) {
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

const testAccOrganizationDataSourceConfig = `
data "catalyst_organization" "test" {}
`
