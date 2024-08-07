package project_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/diagridio/terraform-provider-catalyst/internal/provider/region"
	"github.com/diagridio/terraform-provider-catalyst/internal/test/acceptance"
)

func TestAccProjectDataSource(t *testing.T) {
	t.Skip("skipping")

	name := acctest.RandomWithPrefix("prj")
	// fetch region from environment and default to const if not defined
	region := region.GetEnvOrDefault("TEST_CATALYST_REGION", region.DefaultRegion)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read project
			{
				Config: testAccProjectResourceConfig(name, region.ID(), false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_project.test", "name", name),
					resource.TestCheckResourceAttr("catalyst_project.test", "region", region.ID()),
				),
			},
			// Read project datasource
			{
				Config: testAccProjectDatasourceConfig(name, region.ID()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.catalyst_project.test", "name", name),
					resource.TestCheckResourceAttr("data.catalyst_project.test", "region", region.ID()),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProjectDatasourceConfig(name, region string) string {
	return fmt.Sprintf(`
data "catalyst_organization" "current" {}

data "catalyst_region" "test" {
	id = %q
}

data "catalyst_project" "test" {
  organization_id = data.catalyst_organization.current.id
  region = data.catalyst_region.test.id
  name = %q
}
`, region, name)
}
