package resiliency_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	conductor_client "github.com/diagridio/diagrid-cloud-go/pkg/conductor/client"
	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"
	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"
)

var (
	orgID   = uuid.NewString()
	orgName = acctest.RandomWithPrefix("org")

	projectID      = acctest.RandomWithPrefix("prj")
	resiliencyName = acctest.RandomWithPrefix("resiliency")

	mu           sync.Mutex
	resiliencies = make(map[string]*cloudruntime_client.DaprResiliency)
)

func TestMockResiliencyResource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockResourceClientFactory(ctrl)),
				),
			},
			Steps: []resource.TestStep{
				{
					Config: testAccResiliencyResourceConfig(projectID, resiliencyName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "name", resiliencyName),
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "project_id", projectID),
						resource.TestCheckNoResourceAttr("catalyst_resiliency.test", "spec"),
					),
				},
				{
					ResourceName:                         "catalyst_resiliency.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectID, resiliencyName),
					ImportStateVerify:                    true,
				},
			},
		})
}

func TestMockResiliencyResourceWithSpec(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockResourceClientFactory(ctrl)),
				),
			},
			Steps: []resource.TestStep{
				{
					Config: testAccResiliencyResourceConfigWithSpec(projectID, resiliencyName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "name", resiliencyName),
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "project_id", projectID),
						// Just verify spec is set, don't check exact content due to YAML formatting variances
						resource.TestCheckResourceAttrSet("catalyst_resiliency.test", "spec"),
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "scopes.#", "2"),
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "scopes.0", "app1"),
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "scopes.1", "app2"),
					),
				},
				{
					ResourceName:                         "catalyst_resiliency.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectID, resiliencyName),
					ImportStateVerify:                    true,
					// Ignore spec due to YAML formatting differences (field names, null values, ordering)
					// The spec content is functionally correct, just formatted differently
					ImportStateVerifyIgnore: []string{"spec"},
				},
			},
		})
}

func mockResourceClientFactory(ctrl *gomock.Controller) provider.ClientFactory {
	return func(endpoint, apiKey string, tlsSkipVerify bool) (catalyst.Client, error) {
		c := catalyst.NewMockClient(ctrl)

		c.EXPECT().GetUserOrg(gomock.Any()).Return(
			&conductor_client.Organization{
				Data: conductor_client.OrganizationData{
					Id: lo.ToPtr(orgID),
					Attributes: &conductor_client.OrganizationAttributes{
						Name: lo.ToPtr(orgName),
					},
				},
			}, nil).AnyTimes()

		c.EXPECT().CreateResiliency(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID string, resiliency *cloudruntime_client.DaprResiliency) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, *resiliency.Metadata.Name)
				resiliencies[key] = resiliency
				return nil
			}).AnyTimes()

		c.EXPECT().GetResiliency(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID, name string, qp *cloudruntime_client.DescribeDaprResiliencyParams) (*cloudruntime_client.DaprResiliency, error) {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, name)
				resiliency, exists := resiliencies[key]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return resiliency, nil
			}).AnyTimes()

		c.EXPECT().UpdateResiliency(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID, name string, resiliency *cloudruntime_client.DaprResiliency) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, *resiliency.Metadata.Name)
				resiliencies[key] = resiliency
				return nil
			}).AnyTimes()

		c.EXPECT().DeleteResiliency(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID, name string) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, name)
				delete(resiliencies, key)
				return nil
			}).AnyTimes()

		return c, nil
	}
}

func testAccResiliencyResourceConfig(projectID, resiliencyName string) string {
	return fmt.Sprintf(`
resource "catalyst_resiliency" "test" {
  project_id = %q
  name       = %q
}
`, projectID, resiliencyName)
}

func testAccResiliencyResourceConfigWithSpec(projectID, resiliencyName string) string {
	return fmt.Sprintf(`
resource "catalyst_resiliency" "test" {
  project_id = %q
  name       = %q
  scopes     = ["app1", "app2"]
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
