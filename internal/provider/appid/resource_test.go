package appid_test

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

	projectName = acctest.RandomWithPrefix("prj")
	appidName   = acctest.RandomWithPrefix("appid")

	mu     sync.Mutex
	appids = make(map[string]*cloudruntime_client.AppIdentity)
	projs  = make(map[string]bool)
)

func testSteps() []resource.TestStep {
	return []resource.TestStep{
		// Create and Read testing
		{
			Config: testAccAppIdResourceConfig(projectName, appidName),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_appid.test", "name", appidName),
				resource.TestCheckResourceAttr("catalyst_appid.test", "project_id", projectName),
			),
		},
		// ImportState testing
		{
			ResourceName:                         "catalyst_appid.test",
			ImportState:                          true,
			ImportStateVerifyIdentifierAttribute: "name",
			ImportStateId:                        fmt.Sprintf("%s/%s", projectName, appidName),
			ImportStateVerify:                    true,
		},
		// Delete testing automatically occurs in TestCase
	}
}

func TestMockAppIdResource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockResourceClientFactory(ctrl)),
				),
			},
			Steps: testSteps(),
		})
}

func TestAccAppIdResource(t *testing.T) {
	resource.Test(t,
		resource.TestCase{
			PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps:                    testSteps(),
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
						Status: lo.ToPtr("processing"),
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
			CreateAppId(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName string, appId *cloudruntime_client.AppIdentity) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *appId.Metadata.Name)
				appids[key] = appId
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			GetAppId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, qp *cloudruntime_client.DescribeAppIdentityParams) (*cloudruntime_client.AppIdentity, error) {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				appId, exists := appids[key]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return appId, nil
			}).
			AnyTimes()

		c.EXPECT().
			UpdateAppId(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, appId *cloudruntime_client.AppIdentity) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *appId.Metadata.Name)
				appids[key] = appId
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			DeleteAppId(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				delete(appids, key)
				return nil
			}).
			AnyTimes()

		return c, nil
	}
}

func testAccAppIdResourceConfig(projectName, appidName string) string {
	return fmt.Sprintf(`
resource "catalyst_project" "test" {
  name           = %[1]q
  wait_for_ready = true
}

resource "catalyst_appid" "test" {
  project_id = catalyst_project.test.name
  name       = %[2]q
}
`, projectName, appidName)
}
