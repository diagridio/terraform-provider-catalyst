package project_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"

	"github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	conductor_client "github.com/diagridio/diagrid-cloud-go/pkg/conductor/client"
	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
)

var (
	orgID   = uuid.NewString()
	orgName = acctest.RandomWithPrefix("org")

	regionName     = acctest.RandomWithPrefix("region")
	regionHost     = acctest.RandomWithPrefix("regionHost")
	regionIngress  = fmt.Sprintf("https://*.%s.ingress.diagrid.io:443", regionName)
	regionLocation = "us-west-1"
	regionType     = "public"

	projectName = acctest.RandomWithPrefix("prj")

	mu    sync.Mutex
	projs = make(map[string]bool)

	region *cloudruntime_client.Region
)

func TestMockProjectDataSource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockDatasourceClientFactory(t, ctrl)),
				),
			},
			Steps: []resource.TestStep{
				// Create and Read project
				{
					Config: testAccProjectResourceConfig(projectName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_project.test", "name", projectName),
						resource.TestCheckResourceAttr("catalyst_project.test", "region", regionName),
						resource.TestCheckResourceAttr("catalyst_project.test", "grpc_endpoint", fmt.Sprintf("grpc://grpc-%s.%s", projectName, regionIngress)),
						resource.TestCheckResourceAttr("catalyst_project.test", "http_endpoint", fmt.Sprintf("http://http-%s.%s", projectName, regionIngress)),
					),
				},
				// Read project datasource
				{
					// we need to keep both the project resource/datasource in the same test step otherwise terraform test
					// destroys the project
					Config: testAccProjectResourceConfig(projectName) +
						testAccProjectDatasourceConfig(projectName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.catalyst_project.test", "name", projectName),
						resource.TestCheckResourceAttr("data.catalyst_project.test", "region", regionName),
						resource.TestCheckResourceAttr("data.catalyst_project.test", "grpc_endpoint", fmt.Sprintf("grpc://grpc-%s.%s", projectName, regionIngress)),
						resource.TestCheckResourceAttr("data.catalyst_project.test", "http_endpoint", fmt.Sprintf("http://http-%s.%s", projectName, regionIngress)),
					),
				},
				// Delete testing automatically occurs in TestCase
			},
		})
}

func mockDatasourceClientFactory(t *testing.T, ctrl *gomock.Controller) provider.ClientFactory {
	return func(endpoint, apiKey string) (catalyst.Client, error) {
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
			CreateRegion(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, r *cloudruntime_client.Region) (string, error) {
				region = &cloudruntime_client.Region{
					ApiVersion: lo.ToPtr(catalyst.CatalystDiagridV1Beta1),
					Kind:       lo.ToPtr(catalyst.KindRegion),
					Metadata: &cloudruntime_client.Metadata{
						Name: lo.ToPtr(regionName),
					},
					Spec: &cloudruntime_client.RegionSpec{
						Host:     lo.ToPtr(regionHost),
						Ingress:  lo.ToPtr(regionIngress),
						Location: lo.ToPtr(regionLocation),
						Type:     lo.ToPtr(regionType),
					},
					Status: &cloudruntime_client.RegionStatus{
						Status: &client.ProjectSubResourceStatus{
							Status: lo.ToPtr("ready"),
						},
					},
				}
				return "", nil
			}).
			AnyTimes()

		c.EXPECT().
			GetRegion(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, name string) (*cloudruntime_client.Region, error) {
				if region == nil {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return region, nil
			}).
			AnyTimes()

		c.EXPECT().
			DeleteRegion(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, name string) error {
				region = nil
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			CreateProject(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, project *cloudruntime_client.Project) error {
				// mark the project as created
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
						Region: lo.ToPtr(regionName),
					},
					Status: &cloudruntime_client.ProjectStatus{
						Status: lo.ToPtr("processing"),
						Endpoints: &cloudruntime_client.ProjectStatusEndpoint{
							Grpc: &cloudruntime_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("grpc://grpc-%s.%s", projectName, regionIngress)),
							},
							Http: &cloudruntime_client.ProjectStatusEndpointDetails{
								Url: lo.ToPtr(fmt.Sprintf("http://http-%s.%s", projectName, regionIngress)),
							},
						},
					},
				}, nil
			}).
			AnyTimes()

		c.EXPECT().
			DeleteProject(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string) error {
				// mark the project as deleted
				mu.Lock()
				defer mu.Unlock()
				delete(projs, name)
				return nil
			}).
			AnyTimes()

		return c, nil
	}
}

func testAccProjectDatasourceConfig(name string) string {
	return fmt.Sprintf(`
data "catalyst_project" "test" {
  name = %q
}
`, name)
}
