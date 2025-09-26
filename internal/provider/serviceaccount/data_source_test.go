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
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/samber/lo"
	"go.uber.org/mock/gomock"
)

var (
	dsOrgID   = uuid.NewString()
	dsOrgName = acctest.RandomWithPrefix("org")

	dsServiceAccountName        = acctest.RandomWithPrefix("sa-ds")
	dsServiceAccountDescription = "Test service account description for datasource"
	dsServiceAccountOwner       = "test-owner-ds@example.com"
	dsServiceAccountRole        = "viewer"

	dsMu              sync.Mutex
	dsServiceAccounts = make(map[string]*cloudruntime_client.ServiceAccount)
)

func TestMockServiceAccountDataSource(t *testing.T) {
	ctrl := gomock.NewController(t)

	resource.UnitTest(t,
		resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				provider.ProviderName: providerserver.NewProtocol6WithError(
					provider.New("test").WithClientFactory(mockDatasourceClientFactory(ctrl)),
				),
			},
			Steps: []resource.TestStep{
				// Create service account
				{
					Config: testAccServiceAccountResourceConfigDS(dsServiceAccountName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_service_account.test", "name", dsServiceAccountName),
						resource.TestCheckResourceAttr("catalyst_service_account.test", "description", dsServiceAccountDescription),
						resource.TestCheckResourceAttr("catalyst_service_account.test", "owner", dsServiceAccountOwner),
						resource.TestCheckResourceAttr("catalyst_service_account.test", "role", dsServiceAccountRole),
						resource.TestCheckResourceAttrSet("catalyst_service_account.test", "email"),
					),
				},
				// Read service account datasource by name
				{
					Config: testAccServiceAccountResourceConfigDS(dsServiceAccountName) +
						testAccServiceAccountDatasourceConfigByName(dsServiceAccountName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.catalyst_service_account.test_by_name", "name", dsServiceAccountName),
						resource.TestCheckResourceAttr("data.catalyst_service_account.test_by_name", "description", dsServiceAccountDescription),
						resource.TestCheckResourceAttr("data.catalyst_service_account.test_by_name", "owner", dsServiceAccountOwner),
						resource.TestCheckResourceAttr("data.catalyst_service_account.test_by_name", "role", dsServiceAccountRole),
						resource.TestCheckResourceAttrSet("data.catalyst_service_account.test_by_name", "email"),
					),
				},
				// Delete testing automatically occurs in TestCase
			},
		})
}

func mockDatasourceClientFactory(ctrl *gomock.Controller) provider.ClientFactory {
	return func(endpoint, apiKey string) (catalyst.Client, error) {
		c := catalyst.NewMockClient(ctrl)

		c.EXPECT().
			GetUserOrg(gomock.Any()).
			Return(
				&conductor_client.Organization{
					Data: conductor_client.OrganizationData{
						Id: lo.ToPtr(dsOrgID),
						Attributes: &conductor_client.OrganizationAttributes{
							Name: lo.ToPtr(dsOrgName),
						},
					},
				}, nil).
			AnyTimes()

		c.EXPECT().
			CreateServiceAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, sa *cloudruntime_client.ServiceAccount) error {
				dsMu.Lock()
				defer dsMu.Unlock()

				sa.Status = &cloudruntime_client.ServiceAccountStatus{
					Email:     lo.ToPtr(fmt.Sprintf("%s@service.local", *sa.Metadata.Name)),
					Status:    lo.ToPtr("active"),
					UpdatedAt: lo.ToPtr("2024-01-01T00:00:00Z"),
				}
				dsServiceAccounts[*sa.Metadata.Name] = sa
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			GetServiceAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string) (*cloudruntime_client.ServiceAccount, error) {
				dsMu.Lock()
				defer dsMu.Unlock()

				sa, exists := dsServiceAccounts[name]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				return sa, nil
			}).
			AnyTimes()

		c.EXPECT().
			UpdateServiceAccount(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string, sa *cloudruntime_client.ServiceAccount) error {
				dsMu.Lock()
				defer dsMu.Unlock()

				existing, exists := dsServiceAccounts[name]
				if !exists {
					return diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}

				// Update the spec
				existing.Spec = sa.Spec
				dsServiceAccounts[name] = existing
				return nil
			}).
			AnyTimes()

		c.EXPECT().
			DeleteServiceAccount(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, name string) error {
				dsMu.Lock()
				defer dsMu.Unlock()
				delete(dsServiceAccounts, name)
				return nil
			}).
			AnyTimes()

		return c, nil
	}
}

func testAccServiceAccountDatasourceConfigByName(name string) string {
	return fmt.Sprintf(`
data "catalyst_service_account" "test_by_name" {
  name = %q
}
`, name)
}

func testAccServiceAccountResourceConfigDS(name string) string {
	return fmt.Sprintf(`
resource "catalyst_service_account" "test" {
  name        = %q
  description = %q
  owner       = %q
  role        = %q
}
`, name, dsServiceAccountDescription, dsServiceAccountOwner, dsServiceAccountRole)
}
