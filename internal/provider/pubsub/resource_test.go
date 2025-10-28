package pubsub_test

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

	projectName     = acctest.RandomWithPrefix("prj")
	pubsubName      = acctest.RandomWithPrefix("pubsub")
	componentName   = acctest.RandomWithPrefix("component")
	createComponent = true

	mu      sync.Mutex
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
					ResourceName:                         "catalyst_pubsub.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, pubsubName),
					ImportStateVerify:                    true,
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
resource "catalyst_pubsub" "test" {
  project_name     = %q
  name             = %q
  component_name   = %q
  create_component = %t
}
`, projectName, pubsubName, componentName, createComponent)
}

func testAccPubSubDatasourceConfig(projectName, pubsubName string) string {
	return fmt.Sprintf(`
data "catalyst_pubsub" "test" {
  project_name = %q
  name         = %q
}
`, projectName, pubsubName)
}
