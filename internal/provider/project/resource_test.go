package project_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"testing"

	"github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	conductor_client "github.com/diagridio/diagrid-cloud-go/pkg/conductor/client"
	diagrid_errors "github.com/diagridio/diagrid-cloud-go/pkg/errors"
	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
	"github.com/diagridio/terraform-provider-catalyst/internal/provider"
	"github.com/diagridio/terraform-provider-catalyst/internal/test/acceptance"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"
)

func testSteps() []resource.TestStep {
	return []resource.TestStep{
		// Create and Read testing
		{
			ResourceName: "catalyst_project.test",
			Config:       testAccProjectResourceConfig(projectName),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_project.test", "name", projectName),
				resource.TestCheckResourceAttr("catalyst_project.test", "region", regionName),
			),
		},
		// ImportState testing
		{
			ResourceName:                         "catalyst_project.test",
			ImportState:                          true,
			ImportStateVerifyIdentifierAttribute: "name",
			ImportStateId:                        projectName,
			ImportStateVerify:                    true,
			ImportStateVerifyIgnore:              []string{"wait_for_ready"},
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_project.test", "name", projectName),
				resource.TestCheckResourceAttr("catalyst_project.test", "region", regionName),
			),
		},
		// Update and Read testing
		{
			ResourceName: "catalyst_project.test",
			Config: testAccProjectResourceConfig(projectName) +
				testAccProjectDatasourceConfig(projectName),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_project.test", "name", projectName),
				resource.TestCheckResourceAttr("catalyst_project.test", "region", regionName),
			),
		},
		// Delete testing automatically occurs in TestCase
	}
}

func TestAccProjectResource(t *testing.T) {
	resource.Test(t,
		resource.TestCase{
			PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps:                    testSteps(),
		})
}

func TestMockProjectResource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.Test(t,
		resource.TestCase{
			PreCheck: func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockResourceClientFactory(t, ctrl)),
				),
			},
			Steps: testSteps(),
		})
}

func mockResourceClientFactory(t *testing.T, ctrl *gomock.Controller) provider.ClientFactory {
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

func testAccProjectResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "catalyst_region" "test" {
  name = %q
  ingress = %q
  host = %q
  location = %q
}

resource "catalyst_project" "test" {
  region = catalyst_region.test.name
  name = %q
  wait_for_ready = false
}
`, regionName, regionIngress, regionHost, regionLocation, name)
}
