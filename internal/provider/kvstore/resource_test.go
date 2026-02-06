package kvstore_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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

	projectName     = acctest.RandomWithPrefix("prj")
	kvstoreName     = acctest.RandomWithPrefix("kvstore")
	componentName   = acctest.RandomWithPrefix("component")
	createComponent = true

	mu       sync.Mutex
	kvstores = make(map[string]*catalyst_client.KVStore)
	projs    = make(map[string]bool)
)

func TestMockKVStoreResource(t *testing.T) {
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
					Config: testAccKVStoreResourceConfig(projectName, kvstoreName, componentName, createComponent),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "name", kvstoreName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "component_name", componentName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "create_component", "true"),
					),
				},
				{
					Config: testAccKVStoreResourceConfigWithScopes(projectName, kvstoreName, componentName, createComponent, []string{"app1"}),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "name", kvstoreName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "scopes.0", "app1"),
					),
				},
				{
					ResourceName:                         "catalyst_kvstore.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, kvstoreName),
					ImportStateVerify:                    true,
					ImportStateVerifyIgnore:              []string{"component_name", "create_component", "status"},
				},
			},
		})
}

func TestAccKVStoreResource(t *testing.T) {
	resource.Test(t,
		resource.TestCase{
			PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccKVStoreResourceConfig(projectName, kvstoreName, componentName, createComponent),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "name", kvstoreName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "component_name", componentName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "create_component", "true"),
					),
				},
				{
					ResourceName:                         "catalyst_kvstore.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, kvstoreName),
					ImportStateVerify:                    true,
					ImportStateVerifyIgnore:              []string{"component_name", "create_component", "status"},
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
						Name: lo.ToPtr(projectName),
					},
					Spec: &catalyst_client.ProjectSpec{
						Region: lo.ToPtr("default"),
					},
					Status: &catalyst_client.ProjectStatus{
						Status: lo.ToPtr("ready"),
						Endpoints: &catalyst_client.ProjectStatusEndpoint{
							Grpc: &catalyst_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("grpc://grpc.%s.default.example.com", projectName)),
							},
							Http: &catalyst_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("https://http.%s.default.example.com", projectName)),
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

		c.EXPECT().CreateKVStore(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName string, kvstore *catalyst_client.KVStore) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *kvstore.Metadata.Name)
				kvstores[key] = kvstore
				return nil
			}).AnyTimes()

		c.EXPECT().GetKVStore(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, qp *catalyst_client.DescribeKVStoreParams) (*catalyst_client.KVStore, error) {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				kvstore, exists := kvstores[key]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				// Ensure status is set
				if kvstore.Status == nil {
					kvstore.Status = &catalyst_client.RegionalResourceStatus{}
				}
				if kvstore.Status.Status == nil {
					status := "Ready"
					kvstore.Status.Status = &status
				}
				return kvstore, nil
			}).AnyTimes()

		c.EXPECT().UpdateKVStore(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, kvstore *catalyst_client.KVStore) error {
				mu.Lock()
				defer mu.Unlock()
				// Ensure component fields are NOT sent on updates
				if kvstore.Spec != nil {
					if kvstore.Spec.ComponentName != nil {
						return fmt.Errorf("component_name should not be included in update payload")
					}
					if kvstore.Spec.CreateComponent != nil {
						return fmt.Errorf("create_component should not be included in update payload")
					}
				}
				key := fmt.Sprintf("%s/%s", projectName, *kvstore.Metadata.Name)
				kvstores[key] = kvstore
				return nil
			}).AnyTimes()

		c.EXPECT().DeleteKVStore(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				delete(kvstores, key)
				return nil
			}).AnyTimes()

		return c, nil
	}
}

func testAccKVStoreResourceConfig(projectName, kvstoreName, componentName string, createComponent bool) string {
	return fmt.Sprintf(`
resource "catalyst_project" "test" {
	name           = %[1]q
}

resource "catalyst_kvstore" "test" {
	project_name     = catalyst_project.test.name
	name             = %[2]q
	component_name   = %[3]q
	create_component = %[4]t
}
`, projectName, kvstoreName, componentName, createComponent)
}

func testAccKVStoreResourceConfigWithScopes(projectName, kvstoreName, componentName string, createComponent bool, scopes []string) string {
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

resource "catalyst_kvstore" "test" {
	project_name     = catalyst_project.test.name
	name             = %[2]q
	component_name   = %[3]q
	create_component = %[4]t
	scopes           = %s
}
`, projectName, kvstoreName, componentName, createComponent, scopesStr)
}
