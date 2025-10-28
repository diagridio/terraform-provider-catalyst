package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectName := fmt.Sprintf("tfacc-prj-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccProjectResourceConfig(projectName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_project.test", "name", projectName),
					resource.TestCheckResourceAttrSet("catalyst_project.test", "grpc_endpoint"),
					resource.TestCheckResourceAttrSet("catalyst_project.test", "http_endpoint"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_project.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        projectName,
				// wait_for_ready is not returned by the API
				ImportStateVerifyIgnore: []string{"wait_for_ready"},
			},
		},
	})
}

func testAccProjectResourceConfig(projectName string) string {
	return fmt.Sprintf(`
resource "catalyst_project" "test" {
  name = %[1]q
}
`, projectName)
}
