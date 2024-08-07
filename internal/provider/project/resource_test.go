package project_test

import (
	"fmt"
	"testing"

	"github.com/diagridio/terraform-provider-catalyst/internal/provider/region"
	"github.com/diagridio/terraform-provider-catalyst/internal/test/acceptance"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	name := acctest.RandomWithPrefix("prj")
	// fetch region from environment and default to const if not defined
	region := region.GetEnvOrDefault("TEST_CATALYST_REGION", region.DefaultRegion)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				ResourceName: "catalyst_project.test",
				Config:       testAccProjectResourceConfig(name, region.ID(), true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_project.test", "name", name),
					resource.TestCheckResourceAttr("catalyst_project.test", "region", region.ID()),
					resource.TestCheckResourceAttr("catalyst_project.test", "managed_workflow", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_project.test",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        name,
				ImportStateVerify:                    true,
			},
			// Update and Read testing
			{
				ResourceName: "catalyst_project.test",
				Config:       testAccProjectResourceConfig(name, region.ID(), false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_project.test", "name", name),
					resource.TestCheckResourceAttr("catalyst_project.test", "region", region.ID()),
					resource.TestCheckResourceAttr("catalyst_project.test", "managed_workflow", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProjectResourceConfig(name, region string, managedWorkflow bool) string {
	return fmt.Sprintf(`
data catalyst_organization "current" {}

data catalyst_region "current" {
  id = %q
}

resource "catalyst_project" "test" {
  organization_id = data.catalyst_organization.current.id
  region = data.catalyst_region.current.id
  name = %q

  managed_workflow = %t
}
`, region, name, managedWorkflow)
}
