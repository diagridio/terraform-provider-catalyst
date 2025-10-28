package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResiliencyResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	resiliencyName := fmt.Sprintf("tfacc-res-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResiliencyResourceConfig(projectID, resiliencyName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_resiliency.test", "name", resiliencyName),
					resource.TestCheckResourceAttr("catalyst_resiliency.test", "project_id", projectID),
					resource.TestCheckResourceAttrSet("catalyst_resiliency.test", "spec"),
					resource.TestCheckResourceAttrSet("catalyst_resiliency.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_resiliency.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        fmt.Sprintf("%s/%s", projectID, resiliencyName),
				ImportStateVerifyIgnore:              []string{"spec", "status"},
			},
		},
	})
}

func testAccResiliencyResourceConfig(projectID, resiliencyName string) string {
	return fmt.Sprintf(`
resource "catalyst_resiliency" "test" {
  project_id = %[1]q
  name       = %[2]q
  spec       = <<-EOT
    policies:
      timeouts:
        general: 5s
        important: 60s
      retries:
        retryForever:
          policy: constant
          duration: 5s
          maxRetries: -1
        important:
          policy: exponential
          maxInterval: 60s
          maxRetries: 30
      circuitBreakers:
        simpleCB:
          maxRequests: 1
          timeout: 30s
          trip: consecutiveFailures >= 5
    targets:
      apps:
        app1:
          timeout: general
          retry: retryForever
          circuitBreaker: simpleCB
        app2:
          timeout: important
          retry: important
          circuitBreaker: simpleCB
  EOT
}
`, projectID, resiliencyName)
}
