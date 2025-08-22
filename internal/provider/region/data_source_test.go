package region_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"

	"github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
)

func TestMockRegionDataSource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockDatasourceClientFactory(t, ctrl)),
				),
			},
			Steps: []resource.TestStep{
				// Read testing
				{
					Config: testAccRegionDataSourceConfig(regionName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.catalyst_region.test", "name", regionName),
						resource.TestCheckResourceAttr("data.catalyst_region.test", "host", regionHost),
						resource.TestCheckResourceAttr("data.catalyst_region.test", "ingress", regionIngress),
						resource.TestCheckResourceAttr("data.catalyst_region.test", "location", regionLocation),
						resource.TestCheckResourceAttr("data.catalyst_region.test", "type", regionType),
						resource.TestCheckResourceAttr("data.catalyst_region.test", "connected", "true"),
					),
				},
			},
		})
}

func mockDatasourceClientFactory(t *testing.T, ctrl *gomock.Controller) provider.ClientFactory {
	return func(endpoint, apiKey string) (catalyst.Client, error) {
		c := catalyst.NewMockClient(ctrl)

		c.EXPECT().
			GetRegion(gomock.Any(), gomock.Any()).
			Return(
				&cloudruntime_client.Region{
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
						Connected: lo.ToPtr(true),
					},
				}, nil).
			AnyTimes()

		return c, nil
	}
}

func testAccRegionDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "catalyst_region" "test" {
	name = %q
}
`, name)
}
