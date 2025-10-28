package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAppIdResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	appidName := fmt.Sprintf("tfacc-aid-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAppIdResourceConfig(projectID, appidName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_appid.test", "name", appidName),
					resource.TestCheckResourceAttr("catalyst_appid.test", "project_id", projectID),
					resource.TestCheckResourceAttr("catalyst_appid.test", "protocol", "http"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_appid.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        fmt.Sprintf("%s/%s", projectID, appidName),
				ImportStateVerifyIgnore:              []string{"api_token_revision"},
			},
		},
	})
}

func testAccAppIdResourceConfig(projectID, appidName string) string {
	return fmt.Sprintf(`
resource "catalyst_appid" "test" {
  project_id = %[1]q
  name       = %[2]q
  protocol   = "http"
}
`, projectID, appidName)
}
