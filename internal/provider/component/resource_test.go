package component_test

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

	projectName      = acctest.RandomWithPrefix("prj")
	componentName    = acctest.RandomWithPrefix("component")
	componentType    = "state.redis"
	componentVersion = "v1"

	mu         sync.Mutex
	components = make(map[string]*cloudruntime_client.DaprComponent)
	projs      = make(map[string]bool)
)

func TestMockComponentResource(t *testing.T) {
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
					Config: testAccComponentResourceConfig(projectName, componentName, "state.redis", "v1"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_component.test", "name", componentName),
						resource.TestCheckResourceAttr("catalyst_component.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.type", "state.redis"),
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.version", "v1"),
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.metadata.#", "3"),
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.metadata.0.name", "redisHost"),
					),
				},
				{
					ResourceName:                         "catalyst_component.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, componentName),
					ImportStateVerify:                    true,
					ImportStateVerifyIgnore:              []string{"status"},
				},
			},
		})
}

func TestAccComponentResource(t *testing.T) {
	resource.Test(t,
		resource.TestCase{
			PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccComponentResourceConfig(projectName, componentName, "state.redis", "v1"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_component.test", "name", componentName),
						resource.TestCheckResourceAttr("catalyst_component.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.type", "state.redis"),
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.version", "v1"),
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.metadata.#", "3"),
					),
				},
				{
					ResourceName:                         "catalyst_component.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, componentName),
					ImportStateVerify:                    true,
					ImportStateVerifyIgnore:              []string{"status"},
				},
			},
		})
}

func mockResourceClientFactory(ctrl *gomock.Controller) provider.ClientFactory {
	return func(endpoint, apiKey string, tlsSkipVerify bool) (catalyst.Client, error) {
		c := catalyst.NewMockClient(ctrl)

		c.EXPECT().
			GetUserOrg(gomock.Any()).
			Return(
				&conductor_client.Organization{
					Data: conductor_client.OrganizationData{
						Id: lo.ToPtr(orgID),
						Attributes: &conductor_client.OrganizationAttributes{
							Name: lo.ToPtr(orgName),
						},
					},
				}, nil).
			AnyTimes()

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
						Name: lo.ToPtr(projectName),
					},
					Spec: &cloudruntime_client.ProjectSpec{
						Region: lo.ToPtr("default"),
					},
					Status: &cloudruntime_client.ProjectStatus{
						Status: lo.ToPtr("ready"),
						Endpoints: &cloudruntime_client.ProjectStatusEndpoint{
							Grpc: &cloudruntime_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr("grpc://grpc-test.default.local:443"),
							},
							Http: &cloudruntime_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr("https://http-test.default.local:443"),
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

		c.EXPECT().
			CreateComponent(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName string, component *cloudruntime_client.DaprComponent) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *component.Metadata.Name)
				components[key] = component
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			GetComponent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, qParams *cloudruntime_client.DescribeDaprComponentParams) (*cloudruntime_client.DaprComponent, error) {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				component, exists := components[key]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return component, nil
			}).
			AnyTimes()

		c.EXPECT().
			UpdateComponent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, component *cloudruntime_client.DaprComponent) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *component.Metadata.Name)
				components[key] = component
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			DeleteComponent(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				delete(components, key)
				return nil
			}).
			AnyTimes()

		return c, nil
	}
}

func TestComponentMetadataUpdate(t *testing.T) {
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
					Config: testAccComponentResourceConfig(projectName, componentName, "state.redis", "v1"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.metadata.1.value", "abc123"),
					),
				},
				{
					Config: testAccComponentResourceConfig(
						projectName,
						componentName,
						"state.redis",
						"v1",
						`[
					    {
					      name  = "redisHost"
					      value = "localhost:6379"
					    },
					    {
					      name  = "redisPassword"
					      value = "xyz987"
					    },
					    {
					      name  = "actorStateStore"
					      value = "true"
					    }
					  ]`),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.metadata.1.value", "xyz987"),
					),
				},
			},
		})
}

// TestComponentUpdate verifies that actual spec changes trigger updates correctly
func TestComponentUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockResourceClientFactory(ctrl)),
				),
			},
			Steps: []resource.TestStep{
				// Create
				{
					Config: testAccComponentResourceConfig(projectName, componentName, "state.redis", "v1"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.version", "v1"),
					),
				},
				// Update version
				{
					Config: testAccComponentResourceConfig(projectName, componentName, "state.redis", "v2"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_component.test", "spec.version", "v2"),
					),
				},
			},
		})
}

func testAccComponentResourceConfig(projectName, componentName, componentType, version string, metadataOverride ...string) string {
	metadata := `[
		{
			name  = "redisHost"
			value = "localhost:6379"
		},
		{
			name  = "redisPassword"
			value = "abc123"
		},
		{
			name  = "actorStateStore"
			value = "true"
		}
	]`

	if len(metadataOverride) > 0 {
		metadata = metadataOverride[0]
	}

	return fmt.Sprintf(`
resource "catalyst_project" "test" {
	name           = %q
	wait_for_ready = true
}

resource "catalyst_component" "test" {
	project_name = catalyst_project.test.name
	name         = %q

	spec = {
		type    = %q
		version = %q
		metadata = %s
	}
}
`, projectName, componentName, componentType, version, metadata)
}
