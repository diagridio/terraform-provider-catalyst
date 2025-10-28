package component_test

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

	projectName      = acctest.RandomWithPrefix("prj")
	componentName    = acctest.RandomWithPrefix("component")
	componentType    = "state.redis"
	componentVersion = "v1"

	mu         sync.Mutex
	components = make(map[string]*cloudruntime_client.DaprComponent)
)

func testSteps() []resource.TestStep {
	return []resource.TestStep{
		// Create and Read testing
		{
			Config: testAccComponentResourceConfig(projectName, componentName, componentType, componentVersion),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_component.test", "name", componentName),
				resource.TestCheckResourceAttr("catalyst_component.test", "project_name", projectName),
				resource.TestCheckResourceAttr("catalyst_component.test", "type", componentType),
				resource.TestCheckResourceAttr("catalyst_component.test", "version", componentVersion),
			),
		},
		// ImportState testing
		{
			ResourceName:                         "catalyst_component.test",
			ImportState:                          true,
			ImportStateVerifyIdentifierAttribute: "name",
			ImportStateId:                        fmt.Sprintf("%s/%s", projectName, componentName),
			ImportStateVerify:                    true,
		},
		// Delete testing automatically occurs in TestCase
	}
}

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
						resource.TestCheckResourceAttr("catalyst_component.test", "type", "state.redis"),
						resource.TestCheckResourceAttr("catalyst_component.test", "version", "v1"),
						resource.TestCheckResourceAttrSet("catalyst_component.test", "spec"),
					),
				},
				{
					ResourceName:                         "catalyst_component.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, componentName),
					ImportStateVerify:                    true,
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

func testAccComponentResourceConfig(projectName, componentName, componentType, version string) string {
	return fmt.Sprintf(`
resource "catalyst_component" "test" {
  project_name = %q
  name         = %q
  type         = %q
  version      = %q
  spec         = <<-EOT
    - name: redisHost
      value: localhost:6379
    - name: redisPassword
      secretKeyRef:
        name: redis-secret
        key: password
    - name: actorStateStore
      value: "true"
  EOT
}
`, projectName, componentName, componentType, version)
}
