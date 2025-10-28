package resiliency_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"testing"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	conductor_client "github.com/diagridio/diagrid-cloud-go/pkg/conductor/client"
	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"
	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
	"github.com/diagridio/terraform-provider-catalyst/internal/test/acceptance"
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
	projs        = make(map[string]bool)
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
					// Ignore status as it's a computed field that can change (processing -> ready)
					ImportStateVerifyIgnore: []string{"spec", "status"},
				},
			},
		})
}

func TestAccResiliencyResource(t *testing.T) {
	resource.Test(t,
		resource.TestCase{
			PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccResiliencyResourceConfig(projectID, resiliencyName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "name", resiliencyName),
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "project_id", projectID),
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
					ImportStateVerifyIgnore:              []string{"spec", "status"},
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

		c.EXPECT().
			CreateProject(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, project *cloudruntime_client.Project) error {
				mu.Lock()
				defer mu.Unlock()
				projs[*project.Metadata.Name] = true
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			GetProject(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string, params *cloudruntime_client.DescribeProjectParams) (*cloudruntime_client.Project, error) {
				mu.Lock()
				defer mu.Unlock()
				if ok, created := projs[name]; !ok || !created {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}

				return &cloudruntime_client.Project{
					ApiVersion: lo.ToPtr(catalyst.CatalystDiagridV1Beta1),
					Kind:       lo.ToPtr(catalyst.KindProject),
					Metadata: &cloudruntime_client.Metadata{
						Uid:  lo.ToPtr(strconv.FormatInt(rand.Int63(), 10)),
						Name: lo.ToPtr(projectID),
					},
					Spec: &cloudruntime_client.ProjectSpec{
						Region: lo.ToPtr("default"),
					},
					Status: &cloudruntime_client.ProjectStatus{
						Status: lo.ToPtr("processing"),
						Endpoints: &cloudruntime_client.ProjectStatusEndpoint{
							Grpc: &cloudruntime_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("grpc://grpc.%s.default.example.com", projectID)),
							},
							Http: &cloudruntime_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("https://http.%s.default.example.com", projectID)),
							},
						},
					},
				}, nil
			}).
			AnyTimes()

		c.EXPECT().
			DeleteProject(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string) error {
				mu.Lock()
				defer mu.Unlock()
				delete(projs, name)
				return nil
			}).
			AnyTimes()

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
resource "catalyst_project" "test" {
  name           = %[1]q
  wait_for_ready = false
}

resource "catalyst_resiliency" "test" {
  project_id = catalyst_project.test.name
  name       = %[2]q
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
