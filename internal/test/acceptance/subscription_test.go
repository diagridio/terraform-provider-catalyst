package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSubscriptionResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	subscriptionName := fmt.Sprintf("tfacc-sub-%d", acctest.RandIntRange(1000, 9999))
	pubsubName := fmt.Sprintf("tfacc-ps-%d", acctest.RandIntRange(1000, 9999))
	topicName := fmt.Sprintf("tfacc-topic-%d", acctest.RandIntRange(1000, 9999))
	appidName := fmt.Sprintf("tfacc-aid-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSubscriptionResourceConfig(projectID, subscriptionName, pubsubName, topicName, appidName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_subscription.test", "name", subscriptionName),
					resource.TestCheckResourceAttr("catalyst_subscription.test", "project_name", projectID),
					resource.TestCheckResourceAttr("catalyst_subscription.test", "pubsub_name", pubsubName),
					resource.TestCheckResourceAttr("catalyst_subscription.test", "topic", topicName),
					resource.TestCheckResourceAttrSet("catalyst_subscription.test", "spec"),
					resource.TestCheckResourceAttrSet("catalyst_subscription.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_subscription.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        fmt.Sprintf("%s/%s", projectID, subscriptionName),
				ImportStateVerifyIgnore:              []string{"spec", "status"},
			},
		},
	})
}

func testAccSubscriptionResourceConfig(projectID, subscriptionName, pubsubName, topicName, appidName string) string {
	return fmt.Sprintf(`
resource "catalyst_appid" "test_appid" {
  project_id = %[1]q
  name       = %[5]q
  protocol   = "http"
}

resource "catalyst_pubsub" "test_pubsub" {
  project_name     = %[1]q
  name             = %[3]q
  component_name   = %[3]q
  create_component = true
}

resource "catalyst_subscription" "test" {
  project_name = %[1]q
  name         = %[2]q
  pubsub_name  = catalyst_pubsub.test_pubsub.name
  topic        = %[4]q
  scopes       = [catalyst_appid.test_appid.name]
  spec         = <<-EOT
    routes:
      default: /events
      rules:
        - match: event.type == "order.created"
          path: /orders/new
        - match: event.type == "order.updated"
          path: /orders/update
    bulkSubscribe:
      enabled: true
      maxMessagesCount: 100
      maxAwaitDurationMs: 1000
    metadata:
      maxConcurrentHandlers: "10"
      subscriptionType: fanout
    deadLetterTopic: dlq-topic
  EOT
}
`, projectID, subscriptionName, pubsubName, topicName, appidName)
}
