package pubsub_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

	projectName     = acctest.RandomWithPrefix("prj")
	pubsubName      = acctest.RandomWithPrefix("pubsub")
	componentName   = acctest.RandomWithPrefix("component")
	createComponent = true

	mu      sync.Mutex
	projs   = make(map[string]bool)
	pubsubs = make(map[string]*cloudruntime_client.PubSub)
)

func TestMockPubSubResource(t *testing.T) {
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
					Config: testAccPubSubResourceConfig(projectName, pubsubName, componentName, createComponent),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "name", pubsubName),
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "component_name", componentName),
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "create_component", "true"),
					),
				},
				{
					Config: testAccPubSubResourceConfigWithScopes(projectName, pubsubName, componentName, createComponent, []string{"app1"}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "name", pubsubName),
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "scopes.0", "app1"),
					),
				},
				{
					ResourceName:                         "catalyst_pubsub.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, pubsubName),
					ImportStateVerify:                    true,
					ImportStateVerifyIgnore:              []string{"component_name", "create_component", "status"},
				},
			},
		})
}

func TestAccPubSubResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPubSubResourceConfig(projectName, pubsubName, componentName, createComponent),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_pubsub.test", "name", pubsubName),
					resource.TestCheckResourceAttr("catalyst_pubsub.test", "project_name", projectName),
					resource.TestCheckResourceAttr("catalyst_pubsub.test", "component_name", componentName),
					resource.TestCheckResourceAttr("catalyst_pubsub.test", "create_component", "true"),
				),
			},
			{
				ResourceName:                         "catalyst_pubsub.test",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        fmt.Sprintf("%s/%s", projectName, pubsubName),
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"component_name", "create_component", "status"},
			},
		},
	})
}

func TestMockPubSubDataSource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockResourceClientFactory(ctrl)),
				),
			},
			Steps: []resource.TestStep{
				// Create PubSub resource
				{
					Config: testAccPubSubResourceConfig(projectName, pubsubName, componentName, createComponent),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "name", pubsubName),
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_pubsub.test", "component_name", componentName),
					),
				},
				// Read PubSub datasource
				{
					Config: testAccPubSubResourceConfig(projectName, pubsubName, componentName, createComponent) +
						testAccPubSubDatasourceConfig(projectName, pubsubName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.catalyst_pubsub.test", "name", pubsubName),
						resource.TestCheckResourceAttr("data.catalyst_pubsub.test", "project_name", projectName),
						resource.TestCheckResourceAttr("data.catalyst_pubsub.test", "component_name", componentName),
					),
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

		c.EXPECT().CreateProject(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, project *cloudruntime_client.Project) error {
				mu.Lock()
				defer mu.Unlock()
				projs[*project.Metadata.Name] = true
				return nil
			}).AnyTimes()

		c.EXPECT().GetProject(gomock.Any(), gomock.Any(), gomock.Any()).
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
						Uid:  lo.ToPtr(uuid.NewString()),
						Name: lo.ToPtr(projectName),
					},
					Spec: &cloudruntime_client.ProjectSpec{
						Region: lo.ToPtr("default"),
					},
					Status: &cloudruntime_client.ProjectStatus{
						Status: lo.ToPtr("ready"),
						Endpoints: &cloudruntime_client.ProjectStatusEndpoint{
							Grpc: &cloudruntime_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("grpc://grpc.%s.default.example.com", projectName)),
							},
							Http: &cloudruntime_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("https://http.%s.default.example.com", projectName)),
							},
						},
					},
				}, nil
			}).AnyTimes()

		c.EXPECT().DeleteProject(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string) error {
				mu.Lock()
				defer mu.Unlock()
				delete(projs, name)
				return nil
			}).AnyTimes()

		c.EXPECT().CreatePubSub(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName string, pubsub *cloudruntime_client.PubSub) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *pubsub.Metadata.Name)
				pubsubs[key] = pubsub
				return nil
			}).AnyTimes()

		c.EXPECT().GetPubSub(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, qp *cloudruntime_client.DescribePubSubParams) (*cloudruntime_client.PubSub, error) {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				pubsub, exists := pubsubs[key]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return pubsub, nil
			}).AnyTimes()

		c.EXPECT().UpdatePubSub(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectId, pubsubId string, pubsub *cloudruntime_client.PubSub) error {
				mu.Lock()
				defer mu.Unlock()
				// Ensure component fields are NOT sent on updates
				if pubsub.Spec != nil {
					if pubsub.Spec.ComponentName != nil {
						return fmt.Errorf("component_name should not be included in update payload")
					}
					if pubsub.Spec.CreateComponent != nil {
						return fmt.Errorf("create_component should not be included in update payload")
					}
				}
				key := fmt.Sprintf("%s/%s", projectName, *pubsub.Metadata.Name)
				pubsubs[key] = pubsub
				return nil
			}).AnyTimes()

		c.EXPECT().DeletePubSub(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				delete(pubsubs, key)
				return nil
			}).AnyTimes()

		return c, nil
	}
}

func testAccPubSubResourceConfig(projectName, pubsubName, componentName string, createComponent bool) string {
	return fmt.Sprintf(`
resource "catalyst_project" "test" {
	name           = %[1]q
}

resource "catalyst_pubsub" "test" {
	project_name     = catalyst_project.test.name
	name             = %[2]q
	component_name   = %[3]q
	create_component = %[4]t
}
`, projectName, pubsubName, componentName, createComponent)
}

func testAccPubSubResourceConfigWithScopes(projectName, pubsubName, componentName string, createComponent bool, scopes []string) string {
	scopesStr := "[]"
	if len(scopes) > 0 {
		// build HCL list
		items := make([]string, 0, len(scopes))
		for _, s := range scopes {
			items = append(items, fmt.Sprintf("%q", s))
		}
		scopesStr = fmt.Sprintf("[%s]", strings.Join(items, ","))
	}

	return fmt.Sprintf(`
resource "catalyst_project" "test" {
	name           = %[1]q
}

resource "catalyst_pubsub" "test" {
	project_name     = catalyst_project.test.name
	name             = %[2]q
	component_name   = %[3]q
	create_component = %[4]t
	scopes           = %s
}
`, projectName, pubsubName, componentName, createComponent, scopesStr)
}

func testAccPubSubDatasourceConfig(projectName, pubsubName string) string {
	return fmt.Sprintf(`
data "catalyst_pubsub" "test" {
  project_name = %q
  name         = %q
}
`, projectName, pubsubName)
}
