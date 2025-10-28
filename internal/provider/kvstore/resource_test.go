package kvstore_test

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

	projectName     = acctest.RandomWithPrefix("prj")
	kvstoreName     = acctest.RandomWithPrefix("kvstore")
	componentName   = acctest.RandomWithPrefix("component")
	createComponent = true

	mu       sync.Mutex
	kvstores = make(map[string]*cloudruntime_client.KVStore)
)

func TestMockKVStoreResource(t *testing.T) {
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
					Config: testAccKVStoreResourceConfig(projectName, kvstoreName, componentName, createComponent),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "name", kvstoreName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "project_name", projectName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "component_name", componentName),
						resource.TestCheckResourceAttr("catalyst_kvstore.test", "create_component", "true"),
					),
				},
				{
					ResourceName:                         "catalyst_kvstore.test",
					ImportState:                          true,
					ImportStateVerifyIdentifierAttribute: "name",
					ImportStateId:                        fmt.Sprintf("%s/%s", projectName, kvstoreName),
					ImportStateVerify:                    true,
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

		c.EXPECT().CreateKVStore(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName string, kvstore *cloudruntime_client.KVStore) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *kvstore.Metadata.Name)
				kvstores[key] = kvstore
				return nil
			}).AnyTimes()

		c.EXPECT().GetKVStore(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, qp *cloudruntime_client.DescribeKVStoreParams) (*cloudruntime_client.KVStore, error) {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				kvstore, exists := kvstores[key]
				if !exists {
					return nil, diagrid_errors.NewDiagridCloudError(http.StatusNotFound)
				}
				// Ensure status is set
				if kvstore.Status == nil {
					kvstore.Status = &cloudruntime_client.ProjectSubResourceStatus{}
				}
				if kvstore.Status.Status == nil {
					status := "Ready"
					kvstore.Status.Status = &status
				}
				return kvstore, nil
			}).AnyTimes()

		c.EXPECT().UpdateKVStore(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string, kvstore *cloudruntime_client.KVStore) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, *kvstore.Metadata.Name)
				kvstores[key] = kvstore
				return nil
			}).AnyTimes()

		c.EXPECT().DeleteKVStore(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, projectName, name string) error {
				mu.Lock()
				defer mu.Unlock()
				key := fmt.Sprintf("%s/%s", projectName, name)
				delete(kvstores, key)
				return nil
			}).AnyTimes()

		return c, nil
	}
}

func testAccKVStoreResourceConfig(projectName, kvstoreName, componentName string, createComponent bool) string {
	return fmt.Sprintf(`
resource "catalyst_kvstore" "test" {
  project_name     = %q
  name             = %q
  component_name   = %q
  create_component = %t
}
`, projectName, kvstoreName, componentName, createComponent)
}
