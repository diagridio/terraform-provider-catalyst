package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKVStoreResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	kvstoreName := fmt.Sprintf("tfacc-kv-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKVStoreResourceConfig(projectID, kvstoreName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_kvstore.test", "name", kvstoreName),
					resource.TestCheckResourceAttr("catalyst_kvstore.test", "project_name", projectID),
					resource.TestCheckResourceAttr("catalyst_kvstore.test", "component_name", kvstoreName),
					resource.TestCheckResourceAttrSet("catalyst_kvstore.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_kvstore.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        fmt.Sprintf("%s/%s", projectID, kvstoreName),
				ImportStateVerifyIgnore:              []string{"component_name", "create_component"},
			},
		},
	})
}

func testAccKVStoreResourceConfig(projectID, kvstoreName string) string {
	return fmt.Sprintf(`
resource "catalyst_kvstore" "test" {
  project_name     = %[1]q
  name             = %[2]q
  component_name   = %[2]q
  create_component = true
}
`, projectID, kvstoreName)
}
