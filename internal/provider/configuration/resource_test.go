package configuration_test

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

	projectID         = acctest.RandomWithPrefix("prj")
	configurationName = acctest.RandomWithPrefix("configuration")

	mu             sync.Mutex
	configurations = make(map[string]*catalyst_client.DaprConfiguration)
	projs          = make(map[string]bool)
)

func TestMockConfigurationResource(t *testing.T) {
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
					Config: testAccConfigurationResourceConfig(projectID, configurationName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_configuration.test", "name", configurationName),
						resource.TestCheckResourceAttr("catalyst_configuration.test", "project_id", projectID),
						resource.TestCheckResourceAttr("catalyst_configuration.test", "spec.access_control.default_action", "allow"),
						resource.TestCheckResourceAttr("catalyst_configuration.test", "spec.app_http_pipeline.handlers.#", "1"),
					),
				},
				{
					ResourceName:                         "catalyst_configuration.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectID, configurationName),
					ImportStateVerify:                    true,
					ImportStateVerifyIgnore:              []string{"status", "spec"},
				},
			},
		})
}

func TestAccConfigurationResource(t *testing.T) {
	resource.Test(t,
		resource.TestCase{
			PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigurationResourceConfig(projectID, configurationName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_configuration.test", "name", configurationName),
						resource.TestCheckResourceAttr("catalyst_configuration.test", "project_id", projectID),
						resource.TestCheckResourceAttr("catalyst_configuration.test", "spec.access_control.default_action", "allow"),
						resource.TestCheckResourceAttr("catalyst_configuration.test", "spec.http_pipeline.handlers.#", "1"),
					),
				},
				{
					ResourceName:                         "catalyst_configuration.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectID, configurationName),
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

		c.EXPECT().CreateConfiguration(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID string, configuration *catalyst_client.DaprConfiguration) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, *configuration.Metadata.Name)
				configurations[key] = configuration
				return nil
			}).AnyTimes()

		c.EXPECT().GetConfiguration(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID, name string, qp *catalyst_client.DescribeDaprConfigurationParams) (*catalyst_client.DaprConfiguration, error) {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, name)
				configuration, exists := configurations[key]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return configuration, nil
			}).AnyTimes()

		c.EXPECT().UpdateConfiguration(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID, name string, configuration *catalyst_client.DaprConfiguration) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, *configuration.Metadata.Name)
				configurations[key] = configuration
				return nil
			}).AnyTimes()

		c.EXPECT().DeleteConfiguration(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID, name string) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, name)
				delete(configurations, key)
				return nil
			}).AnyTimes()

		return c, nil
	}
}

func testAccConfigurationResourceConfig(projectID, configurationName string) string {
	return fmt.Sprintf(`
resource "catalyst_project" "test" {
  name           = %[1]q
}

resource "catalyst_configuration" "test" {
	project_id = catalyst_project.test.name
	name       = %[2]q
	spec = {
		access_control = {
			default_action = "allow"
			trust_domain   = "public"
			policies = [
				{
					app_id         = "app1"
					default_action = "allow"
					trust_domain   = "public"
					operations = [
						{
							name       = "op1"
							action     = "allow"
							http_verbs = ["GET", "POST"]
						}
					]
				},
				{
					app_id         = "app2"
					default_action = "deny"
					operations = [
						{
							name       = "op2"
							action     = "deny"
							http_verbs = ["DELETE"]
						}
					]
				}
			]
		}

		app_http_pipeline = {
			handlers = [
				{
					name = "auth"
					type = "middleware"
				}
			]
		}

		http_pipeline = {
			handlers = [
				{
					name = "ingress-auth"
					type = "middleware"
				}
			]
		}
	}
}
`, projectID, configurationName)
}
