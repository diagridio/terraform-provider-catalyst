package serviceaccount_test

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
	"github.com/diagridio/terraform-provider-catalyst/internal/test/acceptance"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"
)

var (
	orgID   = uuid.NewString()
	orgName = acctest.RandomWithPrefix("org")

	serviceAccountName        = acctest.RandomWithPrefix("sa")
	serviceAccountDescription = "Test service account description"
	serviceAccountOwner       = "test-owner@example.com"
	serviceAccountRole        = "cra.diagrid:admin"

	mu              sync.Mutex
	serviceAccounts = make(map[string]*cloudruntime_client.ServiceAccount)
)

func testSteps() []resource.TestStep {
	return []resource.TestStep{
		// Create and Read testing
		{
			ResourceName: "catalyst_service_account.test",
			Config:       testAccServiceAccountResourceConfig(serviceAccountName),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_service_account.test", "name", serviceAccountName),
				resource.TestCheckResourceAttr("catalyst_service_account.test", "description", serviceAccountDescription),
				resource.TestCheckResourceAttr("catalyst_service_account.test", "owner", serviceAccountOwner),
				resource.TestCheckResourceAttr("catalyst_service_account.test", "role", serviceAccountRole),
				resource.TestCheckResourceAttrSet("catalyst_service_account.test", "email"),
			),
		},
		// ImportState testing
		{
			ResourceName:                         "catalyst_service_account.test",
			ImportState:                          true,
			ImportStateVerifyIdentifierAttribute: "name",
			ImportStateIdFunc: func(s *terraform.State) (string, error) {
				return s.RootModule().Resources["catalyst_service_account.test"].Primary.Attributes["name"], nil
			},
			ImportStateVerify: true,
		},
		// Update and Read testing
		{
			ResourceName: "catalyst_service_account.test",
			Config:       testAccServiceAccountResourceConfigUpdated(serviceAccountName),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("catalyst_service_account.test", "name", serviceAccountName),
				resource.TestCheckResourceAttr("catalyst_service_account.test", "description", "Updated description"),
				resource.TestCheckResourceAttr("catalyst_service_account.test", "owner", serviceAccountOwner),
				resource.TestCheckResourceAttr("catalyst_service_account.test", "role", "cra.diagrid:viewer"),
			),
		},
		// Delete testing automatically occurs in TestCase
	}
}

func TestAccServiceAccountResource(t *testing.T) {
	resource.Test(t,
		resource.TestCase{
			PreCheck:                 func() { acceptance.TestAccPreCheck(t) },
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps:                    testSteps(),
		})
}

func TestMockServiceAccountResource(t *testing.T) {
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
			CreateServiceAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, sa *cloudruntime_client.ServiceAccount) error {
				mu.Lock()
				defer mu.Unlock()

				uid := uuid.NewString()
				sa.Metadata.Uid = lo.ToPtr(uid)
				sa.Status = &cloudruntime_client.ServiceAccountStatus{
					Email:     lo.ToPtr(fmt.Sprintf("%s@service.local", *sa.Metadata.Name)),
					Status:    lo.ToPtr("active"),
					UpdatedAt: lo.ToPtr("2024-01-01T00:00:00Z"),
				}
				serviceAccounts[*sa.Metadata.Name] = sa
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			GetServiceAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string) (*cloudruntime_client.ServiceAccount, error) {
				mu.Lock()
				defer mu.Unlock()

				sa, exists := serviceAccounts[name]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return sa, nil
			}).
			AnyTimes()

		c.EXPECT().
			UpdateServiceAccount(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string, sa *cloudruntime_client.ServiceAccount) error {
				mu.Lock()
				defer mu.Unlock()

				existing, exists := serviceAccounts[name]
				if !exists {
					return diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}

				// Update the spec
				existing.Spec = sa.Spec
				serviceAccounts[name] = existing
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			DeleteServiceAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string) error {
				mu.Lock()
				defer mu.Unlock()
				delete(serviceAccounts, name)
				return nil
			}).
			AnyTimes()

		return c, nil
	}
}

func testAccServiceAccountResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "catalyst_service_account" "test" {
  name        = %q
  description = %q
  owner       = %q
  role        = %q
}
`, name, serviceAccountDescription, serviceAccountOwner, serviceAccountRole)
}

func testAccServiceAccountResourceConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "catalyst_service_account" "test" {
  name        = %q
  description = "Updated description"
  owner       = %q
  role        = "cra.diagrid:viewer"
}
`, name, serviceAccountOwner)
}
