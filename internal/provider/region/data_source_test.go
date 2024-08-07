package region_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/diagridio/terraform-provider-catalyst/internal/provider/region"
	"github.com/diagridio/terraform-provider-catalyst/internal/test/acceptance"
)

func TestAccRegionDataSource(t *testing.T) {
	region := region.GetEnvOrDefault("TEST_CATALYST_REGION", region.DefaultRegion)

	resource.Test(t,
		resource.TestCase{
			PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Read testing
				{
					Config: testAccRegionDataSourceConfig(region.ID()),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.catalyst_region.test", "id", region.ID()),
						resource.TestCheckResourceAttr("data.catalyst_region.test", "name", region.Name()),
					),
				},
			},
		})
}

func testAccRegionDataSourceConfig(id string) string {
	return fmt.Sprintf(`
data "catalyst_region" "test" {
	id = %q
}
`, id)
}
