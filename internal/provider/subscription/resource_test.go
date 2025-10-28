package subscription_test

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
	subscriptionName = acctest.RandomWithPrefix("subscription")
	pubsubName       = acctest.RandomWithPrefix("pubsub")
	topicName        = acctest.RandomWithPrefix("topic")

	mu            sync.Mutex
	subscriptions = make(map[string]*cloudruntime_client.DaprSubscription)
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
					Config: testAccSubscriptionResourceConfig(projectName, subscriptionName, pubsubName, topicName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_subscription.test", "name", subscriptionName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "pubsub_name", pubsubName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "topic", topicName),
					),
				},
				{
					ResourceName:                         "catalyst_subscription.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, subscriptionName),
					ImportStateVerify:                    true,
				},
			},
		})
}

func TestMockSubscriptionResourceWithSpec(t *testing.T) {
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
					Config: testAccSubscriptionResourceConfigWithSpec(projectName, subscriptionName, pubsubName, topicName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_subscription.test", "name", subscriptionName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "pubsub_name", pubsubName),
						resource.TestCheckResourceAttr("catalyst_subscription.test", "topic", topicName),
						resource.TestCheckResourceAttrSet("catalyst_subscription.test", "spec"),
					),
				},
				{
					ResourceName:                         "catalyst_subscription.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, subscriptionName),
					ImportStateVerify:                    true,
					// Ignore spec due to YAML formatting differences
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

		c.EXPECT().CreateSubscription(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName string, subscription *cloudruntime_client.DaprSubscription) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *subscription.Metadata.Name)
				subscriptions[key] = subscription
				return nil
			}).AnyTimes()

		c.EXPECT().GetSubscription(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, qp *cloudruntime_client.DescribeDaprSubscriptionParams) (*cloudruntime_client.DaprSubscription, error) {
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
			DoAndReturn(func(ctx context.Context, projectName, name string, subscription *cloudruntime_client.DaprSubscription) error {
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

func testAccSubscriptionResourceConfig(projectName, subscriptionName, pubsubName, topicName string) string {
	return fmt.Sprintf(`
resource "catalyst_subscription" "test" {
  project_name = %q
  name         = %q
  pubsub_name  = %q
  topic        = %q
}
`, projectName, subscriptionName, pubsubName, topicName)
}

func testAccSubscriptionResourceConfigWithSpec(projectName, subscriptionName, pubsubName, topicName string) string {
	return fmt.Sprintf(`
resource "catalyst_subscription" "test" {
  project_name = %q
  name         = %q
  pubsub_name  = %q
  topic        = %q
  spec         = <<-EOT
routes:
    rules:
        - match: event.type == "order.created"
          path: /orders/created
        - match: event.type == "order.updated"
          path: /orders/updated
    default: /orders/default
bulkSubscribe:
    enabled: true
    maxMessagesCount: 100
    maxAwaitDurationMs: 1000
deadLetterTopic: orders-deadletter
metadata:
    maxConcurrentHandlers: "10"
    rawPayload: "true"
  EOT
}
`, projectName, subscriptionName, pubsubName, topicName)
}
