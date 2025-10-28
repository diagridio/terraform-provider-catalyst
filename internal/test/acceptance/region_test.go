package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegionResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	regionName := acctest.RandomWithPrefix("tfacc-region")
	regionHost := acctest.RandomWithPrefix("tfacc-host")
	regionIngress := fmt.Sprintf("https://*.%s.ingress.example.com:443", regionName)
	regionLocation := "us-west-1"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRegionResourceConfig(regionName, regionIngress, regionHost, regionLocation),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_region.test", "name", regionName),
					resource.TestCheckResourceAttr("catalyst_region.test", "ingress", regionIngress),
					resource.TestCheckResourceAttr("catalyst_region.test", "host", regionHost),
					resource.TestCheckResourceAttr("catalyst_region.test", "location", regionLocation),
					resource.TestCheckResourceAttrSet("catalyst_region.test", "join_token"),
				),
			},
			// Note: ImportState testing is skipped due to known issue in provider
			// Update and Read testing
			{
				Config: testAccRegionResourceConfig(regionName, regionIngress, regionHost+"-updated", regionLocation),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_region.test", "name", regionName),
					resource.TestCheckResourceAttr("catalyst_region.test", "host", regionHost+"-updated"),
				),
			},
		},
	})
}

func testAccRegionResourceConfig(regionName, regionIngress, regionHost, regionLocation string) string {
	return fmt.Sprintf(`
resource "catalyst_region" "test" {
  name     = %[1]q
  ingress  = %[2]q
  host     = %[3]q
  location = %[4]q
}
`, regionName, regionIngress, regionHost, regionLocation)
}
