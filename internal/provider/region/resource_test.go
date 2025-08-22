package region_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/samber/lo"
	"go.uber.org/mock/gomock"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
	"github.com/diagridio/terraform-provider-catalyst/internal/test/acceptance"
)

var (
	regionName      = acctest.RandomWithPrefix("region")
	regionHost      = acctest.RandomWithPrefix("regionHost")
	regionIngress   = fmt.Sprintf("https://*.%s.ingress.diagrid.io:443", regionName)
	regionLocation  = "us-west-1"
	regionType      = "public"
	regionJoinToken = acctest.RandomWithPrefix("regionJoinToken")

	region *cloudruntime_client.Region
)

func testSteps() []resource.TestStep {
	return []resource.TestStep{
		// Create and Read testing
		{
			ResourceName: "catalyst_region.test",
			Config:       testAccRegionResourceConfig(regionName, regionIngress, regionHost, regionLocation),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_region.test", "name", regionName),
				resource.TestCheckResourceAttr("catalyst_region.test", "ingress", regionIngress),
				resource.TestCheckResourceAttr("catalyst_region.test", "host", regionHost),
				resource.TestCheckResourceAttr("catalyst_region.test", "location", regionLocation),
				resource.TestCheckResourceAttr("catalyst_region.test", "connected", "false"),
			),
		},
		// DataSource testing
		{
			ResourceName: "data.catalyst_region.test",
			Config: testAccRegionResourceConfig(regionName, regionIngress, regionHost, regionLocation) +
				testAccRegionDataSourceConfig(regionName),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.catalyst_region.test", "name", regionName),
				resource.TestCheckResourceAttr("data.catalyst_region.test", "ingress", regionIngress),
				resource.TestCheckResourceAttr("data.catalyst_region.test", "host", regionHost),
				resource.TestCheckResourceAttr("data.catalyst_region.test", "location", regionLocation),
				resource.TestCheckResourceAttr("data.catalyst_region.test", "connected", "false"),
			),
		},
		// ImportState testing
		{
			ResourceName:                         "catalyst_region.test",
			ImportState:                          true,
			ImportStateVerifyIdentifierAttribute: "name",
			ImportStateId:                        regionName,
			ImportStateVerify:                    true,
			// join token is only ever obtained on creation, so we ignore it in the import state verification
			// since it won't be available
			ImportStateVerifyIgnore: []string{"join_token"},
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_region.test", "name", regionName),
				resource.TestCheckResourceAttr("catalyst_region.test", "ingress", regionIngress),
				resource.TestCheckResourceAttr("catalyst_region.test", "host", regionHost),
				resource.TestCheckResourceAttr("catalyst_region.test", "location", regionLocation),
				resource.TestCheckResourceAttr("catalyst_region.test", "connected", "false"),
			),
		},
		// Update and Read testing
		{
			ResourceName: "catalyst_region.test",
			Config:       testAccRegionResourceConfig(regionName, regionIngress, "regionhost2", "regionLocation2"),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_region.test", "name", regionName),
				resource.TestCheckResourceAttr("catalyst_region.test", "ingress", regionIngress),
				resource.TestCheckResourceAttr("catalyst_region.test", "host", "regionhost2"),
				resource.TestCheckResourceAttr("catalyst_region.test", "location", "regionLocation2"),
				resource.TestCheckResourceAttr("catalyst_region.test", "connected", "false"),
			),
		},
		// DataSource testing after update
		{
			ResourceName: "data.catalyst_region.test",
			Config: testAccRegionResourceConfig(regionName, regionIngress, "regionhost2", "regionLocation2") +
				testAccRegionDataSourceConfig(regionName),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.catalyst_region.test", "name", regionName),
				resource.TestCheckResourceAttr("data.catalyst_region.test", "ingress", regionIngress),
				resource.TestCheckResourceAttr("data.catalyst_region.test", "host", "regionhost2"),
				resource.TestCheckResourceAttr("data.catalyst_region.test", "location", "regionLocation2"),
				resource.TestCheckResourceAttr("data.catalyst_region.test", "connected", "false"),
			),
		},
		// Delete testing automatically occurs in TestCase
	}
}

func TestAccRegionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps:                    testSteps(),
	})
}

func TestMockRegionResource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockResourceClientFactory(t, ctrl)),
				),
			},
			Steps: testSteps(),
		})
}

func mockResourceClientFactory(t *testing.T, ctrl *gomock.Controller) provider.ClientFactory {
	return func(endpoint, apiKey string) (catalyst.Client, error) {
		c := catalyst.NewMockClient(ctrl)

		c.EXPECT().
			GetRegion(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string) (*cloudruntime_client.Region, error) {
				if region == nil {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return region, nil
			}).
			AnyTimes()

		c.EXPECT().
			CreateRegion(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, r *cloudruntime_client.Region) (string, error) {
				region = &cloudruntime_client.Region{
					ApiVersion: lo.ToPtr(catalyst.CatalystDiagridV1Beta1),
					Kind:       lo.ToPtr(catalyst.KindRegion),
					Metadata: &client.Metadata{
						Name: lo.ToPtr(regionName),
					},
					Spec: &client.RegionSpec{
						Host:     lo.ToPtr(regionHost),
						Ingress:  lo.ToPtr(regionIngress),
						Location: lo.ToPtr(regionLocation),
						Type:     lo.ToPtr(regionType),
					},
					Status: &client.RegionStatus{
						Status: &client.ProjectSubResourceStatus{
							Status: lo.ToPtr("ready"),
						},
					},
				}

				return regionJoinToken, nil
			}).
			AnyTimes()

		c.EXPECT().
			UpdateRegion(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, r *cloudruntime_client.Region) (*cloudruntime_client.Region, error) {
				region = r
				region.Spec.Type = lo.ToPtr(regionType)
				return region, nil
			}).
			AnyTimes()

		c.EXPECT().
			DeleteRegion(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string) error {
				region = nil
				return nil
			}).
			AnyTimes()

		return c, nil
	}
}

func testAccRegionResourceConfig(name, ingress, host, location string) string {
	return fmt.Sprintf(`
resource "catalyst_region" "test" {
  name = %q
  ingress = %q
  host = %q
  location = %q
}
`, name, ingress, host, location)
}
