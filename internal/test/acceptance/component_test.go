package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccComponentResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	componentName := fmt.Sprintf("tfacc-cmp-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccComponentResourceConfig(projectID, componentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_component.test", "name", componentName),
					resource.TestCheckResourceAttr("catalyst_component.test", "project_name", projectID),
					resource.TestCheckResourceAttr("catalyst_component.test", "type", "state.redis"),
					resource.TestCheckResourceAttrSet("catalyst_component.test", "spec"),
					resource.TestCheckResourceAttrSet("catalyst_component.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_component.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        fmt.Sprintf("%s/%s", projectID, componentName),
				ImportStateVerifyIgnore:              []string{"spec", "status"},
			},
		},
	})
}

func testAccComponentResourceConfig(projectID, componentName string) string {
	return fmt.Sprintf(`
resource "catalyst_component" "test" {
  project_name = %[1]q
  name         = %[2]q
  type         = "state.redis"
  version      = "v1"
  spec         = <<-EOT
    - name: redisHost
      value: localhost:6379
    - name: redisPassword
      value: testpass123
  EOT
}
`, projectID, componentName)
}
