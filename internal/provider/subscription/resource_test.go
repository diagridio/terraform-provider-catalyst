package subscription_test

import (
	"context"
	"fmt"
	"net/http"
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

	projectName      = acctest.RandomWithPrefix("prj")
	subscriptionName = acctest.RandomWithPrefix("subscription")
	pubsubName       = acctest.RandomWithPrefix("pubsub")
	topicName        = acctest.RandomWithPrefix("topic")
	appidName        = acctest.RandomWithPrefix("appid")

	mu            sync.Mutex
	projs         = make(map[string]bool)
	appids        = make(map[string]*catalyst_client.AppIdentity)
	pubsubs       = make(map[string]*catalyst_client.PubSub)
	subscriptions = make(map[string]*catalyst_client.DaprSubscription)
)

func TestMockSubscriptionResource(t *testing.T) {
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
					Config: testAccSubscriptionResourceConfig(projectName, subscriptionName, pubsubName, topicName, appidName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_subscription.test", "name", subscriptionName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "pubsub_name", pubsubName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "topic", topicName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "spec.routes.default", "/orders/default"),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "spec.routes.rules.#", "2"),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "spec.routes.rules.0.match", "event.type == \"order.created\""),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "spec.bulk_subscribe.enabled", "true"),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "spec.bulk_subscribe.max_messages_count", "100"),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "spec.dead_letter_topic", "orders-deadletter"),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "spec.metadata.maxConcurrentHandlers", "10"),
					),
				},
				{
					ResourceName:                         "catalyst_subscription.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, subscriptionName),
					ImportStateVerify:                    true,
					ImportStateVerifyIgnore:              []string{"status"},
				},
			},
		})
}

func TestAccSubscriptionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionResourceConfig(projectName, subscriptionName, pubsubName, topicName, appidName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_subscription.test", "name", subscriptionName),
					resource.TestCheckResourceAttr("catalyst_subscription.test", "project_name", projectName),
					resource.TestCheckResourceAttr("catalyst_subscription.test", "pubsub_name", pubsubName),
					resource.TestCheckResourceAttr("catalyst_subscription.test", "topic", topicName),
					resource.TestCheckResourceAttr("catalyst_subscription.test", "spec.routes.rules.#", "2"),
					resource.TestCheckResourceAttr("catalyst_subscription.test", "spec.bulk_subscribe.max_await_duration_ms", "1000"),
				),
			},
			{
				ResourceName:                         "catalyst_subscription.test",
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        fmt.Sprintf("%s/%s", projectName, subscriptionName),
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

		c.EXPECT().CreateProject(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, project *catalyst_client.Project) error {
				mu.Lock()
				defer mu.Unlock()
				projs[*project.Metadata.Name] = true
				return nil
			}).AnyTimes()

		c.EXPECT().GetProject(gomock.Any(), gomock.Any(), gomock.Any()).
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
						Uid:  lo.ToPtr(uuid.NewString()),
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
			}).AnyTimes()

		c.EXPECT().DeleteProject(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string) error {
				mu.Lock()
				defer mu.Unlock()
				delete(projs, name)
				return nil
			}).AnyTimes()

		c.EXPECT().CreateAppId(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName string, appId *catalyst_client.AppIdentity) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *appId.Metadata.Name)
				appids[key] = appId
				return nil
			}).AnyTimes()

		c.EXPECT().GetAppId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, qp *catalyst_client.DescribeAppIdentityParams) (*catalyst_client.AppIdentity, error) {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				appId, exists := appids[key]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return appId, nil
			}).AnyTimes()

		c.EXPECT().UpdateAppId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, appId *catalyst_client.AppIdentity) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *appId.Metadata.Name)
				appids[key] = appId
				return nil
			}).AnyTimes()

		c.EXPECT().DeleteAppId(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				delete(appids, key)
				return nil
			}).AnyTimes()

		c.EXPECT().CreatePubSub(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName string, pubsub *catalyst_client.PubSub) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *pubsub.Metadata.Name)
				pubsubs[key] = pubsub
				return nil
			}).AnyTimes()

		c.EXPECT().GetPubSub(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, qp *catalyst_client.DescribePubSubParams) (*catalyst_client.PubSub, error) {
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
			DoAndReturn(func(ctx context.Context, projectId, pubsubId string, pubsub *catalyst_client.PubSub) error {
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

		c.EXPECT().CreateSubscription(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName string, subscription *catalyst_client.DaprSubscription) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *subscription.Metadata.Name)
				subscriptions[key] = subscription
				return nil
			}).AnyTimes()

		c.EXPECT().GetSubscription(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, qp *catalyst_client.DescribeDaprSubscriptionParams) (*catalyst_client.DaprSubscription, error) {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				subscription, exists := subscriptions[key]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return subscription, nil
			}).AnyTimes()

		c.EXPECT().UpdateSubscription(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, subscription *catalyst_client.DaprSubscription) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *subscription.Metadata.Name)
				subscriptions[key] = subscription
				return nil
			}).AnyTimes()

		c.EXPECT().DeleteSubscription(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				delete(subscriptions, key)
				return nil
			}).AnyTimes()

		return c, nil
	}
}

func testAccSubscriptionResourceConfig(projectName, subscriptionName, pubsubName, topicName, appidName string) string {
	return fmt.Sprintf(`
resource "catalyst_project" "test" {
  name           = %[1]q
}

resource "catalyst_appid" "test_appid" {
  project_id = catalyst_project.test.name
  name       = %[5]q
  protocol   = "http"
}

resource "catalyst_pubsub" "test_pubsub" {
  project_name     = catalyst_project.test.name
  name             = %[3]q
  component_name   = %[3]q
  create_component = true
}

resource "catalyst_subscription" "test" {
  project_name = catalyst_project.test.name
  name         = %[2]q
  pubsub_name  = catalyst_pubsub.test_pubsub.name
  topic        = %[4]q
  scopes       = [catalyst_appid.test_appid.name]
	spec = {
		routes = {
			rules = [
				{
					match = "event.type == \"order.created\""
					path  = "/orders/created"
				},
				{
					match = "event.type == \"order.updated\""
					path  = "/orders/updated"
				}
			]
			default = "/orders/default"
		}

		bulk_subscribe = {
			enabled               = true
			max_messages_count    = 100
			max_await_duration_ms = 1000
		}

		dead_letter_topic = "orders-deadletter"

		metadata = {
			maxConcurrentHandlers = "10"
			rawPayload            = "true"
		}
	}
}
`, projectName, subscriptionName, pubsubName, topicName, appidName)
}
