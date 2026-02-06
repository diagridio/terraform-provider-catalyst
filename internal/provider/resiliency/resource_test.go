package resiliency_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"testing"

	catalyst_client "github.com/diagridio/cloudgrid/sdk/go/pkg/catalyst/client"
	conductor_client "github.com/diagridio/cloudgrid/sdk/go/pkg/conductor/client"
	diagrid_errors "github.com/diagridio/cloudgrid/sdk/go/pkg/errors"
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
	resiliencies = make(map[string]*catalyst_client.DaprResiliency)
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
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "spec.policies.timeouts.general", "5s"),
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "spec.policies.retries.retryForever.max_retries", "-1"),
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "spec.targets.apps.app1.circuit_breaker", "simpleCB"),
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
					ImportStateVerifyIgnore:              []string{"status", "spec"},
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
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "spec.policies.timeouts.general", "5s"),
						resource.TestCheckResourceAttr("catalyst_resiliency.test", "spec.targets.apps.app2.retry", "important"),
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
					ImportStateVerifyIgnore:              []string{"status", "spec"},
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
			DoAndReturn(func(ctx context.Context, project *catalyst_client.Project) error {
				mu.Lock()
				defer mu.Unlock()
				projs[*project.Metadata.Name] = true
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			GetProject(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string, params *catalyst_client.DescribeProjectParams) (*catalyst_client.Project, error) {
				mu.Lock()
				defer mu.Unlock()
				if ok, created := projs[name]; !ok || !created {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}

				return &catalyst_client.Project{
					ApiVersion: lo.ToPtr(catalyst.CatalystDiagridV1Beta1),
					Kind:       lo.ToPtr(catalyst.KindProject),
					Metadata: &catalyst_client.Metadata{
						Uid:  lo.ToPtr(strconv.FormatInt(rand.Int63(), 10)),
						Name: lo.ToPtr(projectID),
					},
					Spec: &catalyst_client.ProjectSpec{
						Region: lo.ToPtr("default"),
					},
					Status: &catalyst_client.ProjectStatus{
						Status: lo.ToPtr("ready"),
						Endpoints: &catalyst_client.ProjectStatusEndpoint{
							Grpc: &catalyst_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("grpc://grpc.%s.default.example.com", projectID)),
							},
							Http: &catalyst_client.ProjectStatusEndpointDetails{
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
			DoAndReturn(func(ctx context.Context, projectID string, resiliency *catalyst_client.DaprResiliency) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, *resiliency.Metadata.Name)
				resiliencies[key] = resiliency
				return nil
			}).AnyTimes()

		c.EXPECT().GetResiliency(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID, name string, qp *catalyst_client.DescribeDaprResiliencyParams) (*catalyst_client.DaprResiliency, error) {
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
			DoAndReturn(func(ctx context.Context, projectID, name string, resiliency *catalyst_client.DaprResiliency) error {
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
}

resource "catalyst_resiliency" "test" {
	project_id = catalyst_project.test.name
	name       = %[2]q
	scopes     = ["app1", "app2"]
	spec = {
		policies = {
			timeouts = {
				general   = "5s"
				important = "60s"
			}

			retries = {
				retryForever = {
					policy      = "constant"
					duration    = "5s"
					max_retries = -1
				}

				important = {
					policy       = "exponential"
					max_interval = "60s"
					max_retries  = 30
				}
			}

			circuit_breakers = {
				simpleCB = {
					max_requests = 1
					timeout      = "30s"
					trip         = "consecutiveFailures >= 5"
				}
			}
		}

		targets = {
			apps = {
				app1 = {
					timeout         = "general"
					retry           = "retryForever"
					circuit_breaker = "simpleCB"
				}

				app2 = {
					timeout         = "important"
					retry           = "important"
					circuit_breaker = "simpleCB"
				}
			}
		}
	}
}
`, projectID, resiliencyName)
}
