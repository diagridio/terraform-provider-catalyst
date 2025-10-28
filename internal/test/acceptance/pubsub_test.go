package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPubSubResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	pubsubName := fmt.Sprintf("tfacc-ps-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPubSubResourceConfig(projectID, pubsubName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_pubsub.test", "name", pubsubName),
					resource.TestCheckResourceAttr("catalyst_pubsub.test", "project_name", projectID),
					resource.TestCheckResourceAttr("catalyst_pubsub.test", "component_name", pubsubName),
					resource.TestCheckResourceAttrSet("catalyst_pubsub.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_pubsub.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        fmt.Sprintf("%s/%s", projectID, pubsubName),
				ImportStateVerifyIgnore:              []string{"component_name", "create_component"},
			},
		},
	})
}

func testAccPubSubResourceConfig(projectID, pubsubName string) string {
	return fmt.Sprintf(`
resource "catalyst_pubsub" "test" {
  project_name     = %[1]q
  name             = %[2]q
  component_name   = %[2]q
  create_component = true
}
`, projectID, pubsubName)
}
