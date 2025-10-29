package configuration_test

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

	projectID         = acctest.RandomWithPrefix("prj")
	configurationName = acctest.RandomWithPrefix("configuration")

	mu             sync.Mutex
	configurations = make(map[string]*cloudruntime_client.DaprConfiguration)
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
						resource.TestCheckResourceAttrSet("catalyst_configuration.test", "spec"),
					),
				},
				{
					ResourceName:                         "catalyst_configuration.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectID, configurationName),
					ImportStateVerify:                    true,
					ImportStateVerifyIgnore:              []string{"spec", "status"},
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
						resource.TestCheckResourceAttrSet("catalyst_configuration.test", "spec"),
					),
				},
				{
					ResourceName:                         "catalyst_configuration.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectID, configurationName),
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
						Name: lo.ToPtr(projectID),
					},
					Spec: &cloudruntime_client.ProjectSpec{
						Region: lo.ToPtr("default"),
					},
					Status: &cloudruntime_client.ProjectStatus{
						Status: lo.ToPtr("processing"),
						Endpoints: &cloudruntime_client.ProjectStatusEndpoint{
							Grpc: &cloudruntime_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("grpc://grpc.%s.default.example.com", projectID)),
							},
							Http: &cloudruntime_client.ProjectStatusEndpointDetails{
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
			DoAndReturn(func(ctx context.Context, projectID string, configuration *cloudruntime_client.DaprConfiguration) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectID, *configuration.Metadata.Name)
				configurations[key] = configuration
				return nil
			}).AnyTimes()

		c.EXPECT().GetConfiguration(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectID, name string, qp *cloudruntime_client.DescribeDaprConfigurationParams) (*cloudruntime_client.DaprConfiguration, error) {
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
			DoAndReturn(func(ctx context.Context, projectID, name string, configuration *cloudruntime_client.DaprConfiguration) error {
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
  wait_for_ready = true
}

resource "catalyst_configuration" "test" {
  project_id = catalyst_project.test.name
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
