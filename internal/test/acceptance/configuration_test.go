package acceptance

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigurationResource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	TestAccPreCheck(t)

	projectID := os.Getenv("CATALYST_PROJECT_ID")
	if projectID == "" {
		t.Fatal("CATALYST_PROJECT_ID must be set for acceptance tests")
	}

	configurationName := fmt.Sprintf("tfacc-cfg-%d", acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccConfigurationResourceConfig(projectID, configurationName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("catalyst_configuration.test", "name", configurationName),
					resource.TestCheckResourceAttr("catalyst_configuration.test", "project_id", projectID),
					resource.TestCheckResourceAttrSet("catalyst_configuration.test", "spec"),
					resource.TestCheckResourceAttrSet("catalyst_configuration.test", "status"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "catalyst_configuration.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
				ImportStateId:                        fmt.Sprintf("%s/%s", projectID, configurationName),
				ImportStateVerifyIgnore:              []string{"spec", "status"},
			},
		},
	})
}

func testAccConfigurationResourceConfig(projectID, configurationName string) string {
	return fmt.Sprintf(`
resource "catalyst_configuration" "test" {
  project_id = %[1]q
  name       = %[2]q
  spec       = <<-EOT
    accessControl:
      defaultAction: allow
      trustDomain: public
      policies:
        - appId: app1
          defaultAction: allow
          trustDomain: public
          namespace: default
          operations:
            - name: op1
              httpVerb:
                - GET
                - POST
              action: allow
        - appId: app2
          defaultAction: deny
          operations:
            - name: op2
              httpVerb:
                - DELETE
              action: deny
    api:
      allowed:
        - name: state
          version: v1
          protocol: http
        - name: pubsub
          version: v1
          protocol: grpc
    tracing:
      samplingRate: "1"
      stdout: true
      zipkin:
        endpointAddress: http://zipkin:9411/api/v2/spans
    metrics:
      enabled: true
      http:
        increasedCardinality: true
        pathMatching:
          - /orders/*
          - /products/*
    secrets:
      scopes:
        - storeName: vault
          defaultAccess: allow
          allowedSecrets:
            - secret1
            - secret2
          deniedSecrets:
            - secret3
  EOT
}
`, projectID, configurationName)
}
