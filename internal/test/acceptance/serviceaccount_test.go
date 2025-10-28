package acceptance

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceAccountResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	serviceAccountName := acctest.RandomWithPrefix("tfacc-sa")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccServiceAccountResourceConfig(serviceAccountName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_service_account.test", "name", serviceAccountName),
					resource.TestCheckResourceAttr("catalyst_service_account.test", "owner", "org"),
					resource.TestCheckResourceAttr("catalyst_service_account.test", "role", "cra.diagrid:admin"),
					resource.TestCheckResourceAttrSet("catalyst_service_account.test", "email"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_service_account.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        serviceAccountName,
			},
		},
	})
}

func testAccServiceAccountResourceConfig(serviceAccountName string) string {
	return `
resource "catalyst_service_account" "test" {
  name        = "` + serviceAccountName + `"
  owner       = "org"
  role        = "cra.diagrid:admin"
  description = "Acceptance test service account"
}
`
}
