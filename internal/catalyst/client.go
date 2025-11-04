package catalyst

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"k8s.io/utils/ptr"

	"github.com/diagridio/diagrid-cloud-go/cloudruntime"
	"github.com/diagridio/diagrid-cloud-go/management"
	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	conductor_client "github.com/diagridio/diagrid-cloud-go/pkg/conductor/client"
)

type Client interface {
	GetUserOrg(context.Context) (*conductor_client.Organization, error)

	// Region operations
	CreateRegion(ctx context.Context, region *cloudruntime_client.Region) (string, error)
	GetRegion(ctx context.Context, name string) (*cloudruntime_client.Region, error)
	UpdateRegion(ctx context.Context, region *cloudruntime_client.Region) error
	DeleteRegion(ctx context.Context, name string) error

	// Project operations
	GetProject(ctx context.Context, id string, qp *cloudruntime_client.DescribeProjectParams) (*cloudruntime_client.Project, error)
	CreateProject(ctx context.Context, project *cloudruntime_client.Project) error
	UpdateProject(ctx context.Context, prj *cloudruntime_client.Project) error
	DeleteProject(ctx context.Context, id string) error

	// ServiceAccount operations
	CreateServiceAccount(ctx context.Context, serviceAccount *cloudruntime_client.ServiceAccount) error
	GetServiceAccount(ctx context.Context, serviceAccountId string) (*cloudruntime_client.ServiceAccount, error)
	UpdateServiceAccount(ctx context.Context, serviceAccountId string, serviceAccount *cloudruntime_client.ServiceAccount) error
	DeleteServiceAccount(ctx context.Context, serviceAccountId string) error

	// AppId operations
	CreateAppId(ctx context.Context, projectId string, appid *cloudruntime_client.AppIdentity) error
	GetAppId(ctx context.Context, projectId string, appId string, qp *cloudruntime_client.DescribeAppIdentityParams) (*cloudruntime_client.AppIdentity, error)
	UpdateAppId(ctx context.Context, projectId string, appId string, appid *cloudruntime_client.AppIdentity) error
	DeleteAppId(ctx context.Context, projectId string, appId string) error

	// Component operations
	CreateComponent(ctx context.Context, projectName string, component *cloudruntime_client.DaprComponent) error
	GetComponent(ctx context.Context, projectName string, name string, qParams *cloudruntime_client.DescribeDaprComponentParams) (*cloudruntime_client.DaprComponent, error)
	UpdateComponent(ctx context.Context, projectName string, name string, component *cloudruntime_client.DaprComponent) error
	DeleteComponent(ctx context.Context, projectName string, name string) error

	// PubSub operations
	CreatePubSub(ctx context.Context, projectName string, pubsub *cloudruntime_client.PubSub) error
	GetPubSub(ctx context.Context, projectName string, pubsubId string, qp *cloudruntime_client.DescribePubSubParams) (*cloudruntime_client.PubSub, error)
	UpdatePubSub(ctx context.Context, projectId string, pubsubId string, pubsub *cloudruntime_client.PubSub) error
	DeletePubSub(ctx context.Context, projectId string, pubSubId string) error

	// KVStore operations
	CreateKVStore(ctx context.Context, projectName string, kvstore *cloudruntime_client.KVStore) error
	GetKVStore(ctx context.Context, projectName string, kvStoreName string, qp *cloudruntime_client.DescribeKVStoreParams) (*cloudruntime_client.KVStore, error)
	UpdateKVStore(ctx context.Context, projectName string, kvStoreName string, kvstore *cloudruntime_client.KVStore) error
	DeleteKVStore(ctx context.Context, projectName string, kvStoreName string) error

	// Subscription operations
	CreateSubscription(ctx context.Context, projectID string, subscription *cloudruntime_client.DaprSubscription) error
	GetSubscription(ctx context.Context, projectName string, subscriptionName string, qp *cloudruntime_client.DescribeDaprSubscriptionParams) (*cloudruntime_client.DaprSubscription, error)
	UpdateSubscription(ctx context.Context, projectName string, subscriptionName string, subscription *cloudruntime_client.DaprSubscription) error
	DeleteSubscription(ctx context.Context, projectName string, subscriptionName string) error

	// Resiliency operations
	CreateResiliency(ctx context.Context, projectID string, resiliency *cloudruntime_client.DaprResiliency) error
	GetResiliency(ctx context.Context, projectName string, resiliencyName string, qp *cloudruntime_client.DescribeDaprResiliencyParams) (*cloudruntime_client.DaprResiliency, error)
	UpdateResiliency(ctx context.Context, projectID string, resiliencyName string, resiliency *cloudruntime_client.DaprResiliency) error
	DeleteResiliency(ctx context.Context, projectID string, resiliencyName string) error

	// Configuration operations
	CreateConfiguration(ctx context.Context, projectId string, config *cloudruntime_client.DaprConfiguration) error
	GetConfiguration(ctx context.Context, projectId string, configName string, qp *cloudruntime_client.DescribeDaprConfigurationParams) (*cloudruntime_client.DaprConfiguration, error)
	UpdateConfiguration(ctx context.Context, projectId string, configName string, config *cloudruntime_client.DaprConfiguration) error
	DeleteConfiguration(ctx context.Context, projectId string, configName string) error
}

type cclient struct {
	management management.ManagementClient
	catalyst   cloudruntime.CloudruntimeAPIClient
}

var (
	ErrAPIKeyNotFound   = fmt.Errorf("API key not found in environment variable CATALYST_API_KEY or provider configuration block api_key attribute")
	ErrEndpointNotFound = fmt.Errorf("endpoint not found in environment variable CATALYST_API_ENDPOINT or provider configuration block endpoint attribute")
)

func NewClient(endpoint, apiKey string, tlsSkipVerify bool) (Client, error) {
	if apiKey == "" {
		return nil, ErrAPIKeyNotFound
	}
	if endpoint == "" {
		return nil, ErrEndpointNotFound
	}

	// Create HTTP client with optional TLS skip verify
	httpClient := http.DefaultClient
	if tlsSkipVerify {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		httpClient = &http.Client{
			Transport: transport,
		}
	}

	// Example client configuration for data sources and resources
	maxRetries := 1
	mc, err := management.NewManagementClientWithExponentialBackoff(httpClient,
		endpoint,
		maxRetries,
		management.WithAPIKeyToken(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating management client: %w", err)
	}

	catalystClient, err := cloudruntime.NewCloudruntimeClientWithExponentialBackoff(httpClient,
		endpoint,
		maxRetries,
		cloudruntime.WithAPIKeyToken(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating catalyst client: %w", err)
	}

	return &cclient{
		management: mc,
		catalyst:   catalystClient,
	}, nil
}

func (c *cclient) GetUserOrg(ctx context.Context) (*conductor_client.Organization, error) {
	// find the current user's organization id
	user, err := c.management.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user org: %w", err)
	}

	// now fetch the org
	org, err := c.management.GetUserOrg(ctx, *user.Data.Attributes.Organization.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting user org %s: %w",
			*user.Data.Attributes.Organization.Id, err)
	}

	return org, nil
}

func (c *cclient) CreateRegion(ctx context.Context, region *cloudruntime_client.Region) (string, error) {
	resp, err := c.catalyst.CreatePrivateRegion(ctx, region)
	if err != nil {
		return "", fmt.Errorf("error creating region: %w", err)
	}
	if resp == nil || resp.JoinToken == nil || *resp.JoinToken == "" {
		return "", fmt.Errorf("error creating region: join token is empty")
	}

	return *resp.JoinToken, nil
}

func (c *cclient) GetRegion(ctx context.Context, name string) (*cloudruntime_client.Region, error) {
	region, err := c.catalyst.GetRegion(ctx, name)
	if err != nil {
		return nil, err
	}

	return region, nil
}

func (c *cclient) UpdateRegion(ctx context.Context, region *cloudruntime_client.Region) error {
	if err := c.catalyst.PutPrivateRegion(ctx, *region.Metadata.Name, region); err != nil {
		return fmt.Errorf("error updating region %s: %w", *region.Metadata.Name, err)
	}
	return nil
}

func (c *cclient) DeleteRegion(ctx context.Context, name string) error {
	if err := c.catalyst.DeletePrivateRegion(ctx, name); err != nil {
		return fmt.Errorf("error deleting region %s: %w", name, err)
	}
	return nil
}

func (c *cclient) GetProject(ctx context.Context, id string, qp *cloudruntime_client.DescribeProjectParams) (*cloudruntime_client.Project, error) {
	project, err := c.catalyst.GetProject(ctx, id, qp)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (c *cclient) CreateProject(ctx context.Context, project *cloudruntime_client.Project) error {
	if err := c.catalyst.CreateProject(ctx, project, &cloudruntime_client.CreateProjectParams{
		// We always set wait for ready to true as this ensures we
		// can more safely create sub-resources immediately after.
		WaitForReady: ptr.To(true),
	}); err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	return nil
}

func (c *cclient) UpdateProject(ctx context.Context, project *cloudruntime_client.Project) error {
	if err := c.catalyst.PatchProject(ctx, project); err != nil {
		return fmt.Errorf("error patching project: %w", err)
	}

	return nil
}

func (c *cclient) DeleteProject(ctx context.Context, id string) error {
	if err := c.catalyst.DeleteProject(ctx, id); err != nil {
		return fmt.Errorf("error deleting project %s: %w", id, err)
	}

	return nil
}

func (c *cclient) CreateServiceAccount(ctx context.Context, serviceAccount *cloudruntime_client.ServiceAccount) error {
	if err := c.catalyst.CreateServiceAccount(ctx, serviceAccount); err != nil {
		return fmt.Errorf("error creating service account: %w", err)
	}

	return nil
}

func (c *cclient) GetServiceAccount(ctx context.Context, serviceAccountId string) (*cloudruntime_client.ServiceAccount, error) {
	serviceAccount, err := c.catalyst.GetServiceAccount(ctx, serviceAccountId)
	if err != nil {
		return nil, err
	}

	return serviceAccount, nil
}

func (c *cclient) UpdateServiceAccount(ctx context.Context, serviceAccountId string, serviceAccount *cloudruntime_client.ServiceAccount) error {
	if err := c.catalyst.PatchServiceAccount(ctx, serviceAccountId, serviceAccount); err != nil {
		return fmt.Errorf("error updating service account %s: %w", serviceAccountId, err)
	}

	return nil
}

func (c *cclient) DeleteServiceAccount(ctx context.Context, serviceAccountId string) error {
	if err := c.catalyst.DeleteServiceAccount(ctx, serviceAccountId); err != nil {
		return fmt.Errorf("error deleting service account %s: %w", serviceAccountId, err)
	}

	return nil
}

// AppId operations.
func (c *cclient) CreateAppId(ctx context.Context, projectId string, appid *cloudruntime_client.AppIdentity) error {
	if err := c.catalyst.CreateAppId(ctx, projectId, appid); err != nil {
		return fmt.Errorf("error creating appid: %w", err)
	}
	return nil
}

func (c *cclient) GetAppId(ctx context.Context, projectId string, appId string, qp *cloudruntime_client.DescribeAppIdentityParams) (*cloudruntime_client.AppIdentity, error) {
	appid, err := c.catalyst.GetAppId(ctx, projectId, appId, qp)
	if err != nil {
		return nil, err
	}
	return appid, nil
}

func (c *cclient) UpdateAppId(ctx context.Context, projectId string, appId string, appid *cloudruntime_client.AppIdentity) error {
	if err := c.catalyst.PatchAppId(ctx, projectId, appId, appid); err != nil {
		return fmt.Errorf("error updating appid %s: %w", appId, err)
	}
	return nil
}

func (c *cclient) DeleteAppId(ctx context.Context, projectId string, appId string) error {
	if err := c.catalyst.DeleteAppId(ctx, projectId, appId); err != nil {
		return fmt.Errorf("error deleting appid %s: %w", appId, err)
	}
	return nil
}

// Component operations.
func (c *cclient) CreateComponent(ctx context.Context, projectName string, component *cloudruntime_client.DaprComponent) error {
	if err := c.catalyst.CreateComponent(ctx, projectName, component); err != nil {
		return fmt.Errorf("error creating component: %w", err)
	}
	return nil
}

func (c *cclient) GetComponent(ctx context.Context, projectName string, name string, qParams *cloudruntime_client.DescribeDaprComponentParams) (*cloudruntime_client.DaprComponent, error) {
	component, err := c.catalyst.GetComponent(ctx, projectName, name, qParams)
	if err != nil {
		return nil, err
	}
	return component, nil
}

func (c *cclient) UpdateComponent(ctx context.Context, projectName string, name string, component *cloudruntime_client.DaprComponent) error {
	if err := c.catalyst.PatchComponent(ctx, projectName, name, component); err != nil {
		return fmt.Errorf("error updating component %s: %w", name, err)
	}
	return nil
}

func (c *cclient) DeleteComponent(ctx context.Context, projectName string, name string) error {
	if err := c.catalyst.DeleteComponent(ctx, projectName, name); err != nil {
		return fmt.Errorf("error deleting component %s: %w", name, err)
	}
	return nil
}

// PubSub operations.
func (c *cclient) CreatePubSub(ctx context.Context, projectName string, pubsub *cloudruntime_client.PubSub) error {
	if err := c.catalyst.CreatePubSub(ctx, projectName, pubsub); err != nil {
		return fmt.Errorf("error creating pubsub: %w", err)
	}
	return nil
}

func (c *cclient) GetPubSub(ctx context.Context, projectName string, pubsubId string, qp *cloudruntime_client.DescribePubSubParams) (*cloudruntime_client.PubSub, error) {
	pubsub, err := c.catalyst.GetPubSub(ctx, projectName, pubsubId, qp)
	if err != nil {
		return nil, err
	}
	return pubsub, nil
}

func (c *cclient) UpdatePubSub(ctx context.Context, projectId string, pubsubId string, pubsub *cloudruntime_client.PubSub) error {
	if err := c.catalyst.PatchPubSub(ctx, projectId, pubsubId, pubsub); err != nil {
		return fmt.Errorf("error updating pubsub %s: %w", pubsubId, err)
	}
	return nil
}

func (c *cclient) DeletePubSub(ctx context.Context, projectId string, pubSubId string) error {
	if err := c.catalyst.DeletePubSub(ctx, projectId, pubSubId); err != nil {
		return fmt.Errorf("error deleting pubsub %s: %w", pubSubId, err)
	}
	return nil
}

// KVStore operations.
func (c *cclient) CreateKVStore(ctx context.Context, projectName string, kvstore *cloudruntime_client.KVStore) error {
	if err := c.catalyst.CreateKVStore(ctx, projectName, kvstore); err != nil {
		return fmt.Errorf("error creating kvstore: %w", err)
	}
	return nil
}

func (c *cclient) GetKVStore(ctx context.Context, projectName string, kvStoreName string, qp *cloudruntime_client.DescribeKVStoreParams) (*cloudruntime_client.KVStore, error) {
	kvstore, err := c.catalyst.GetKVStore(ctx, projectName, kvStoreName, qp)
	if err != nil {
		return nil, err
	}
	return kvstore, nil
}

func (c *cclient) UpdateKVStore(ctx context.Context, projectName string, kvStoreName string, kvstore *cloudruntime_client.KVStore) error {
	if err := c.catalyst.PatchKVStore(ctx, projectName, kvStoreName, kvstore); err != nil {
		return fmt.Errorf("error updating kvstore %s: %w", kvStoreName, err)
	}
	return nil
}

func (c *cclient) DeleteKVStore(ctx context.Context, projectName string, kvStoreName string) error {
	if err := c.catalyst.DeleteKVStore(ctx, projectName, kvStoreName); err != nil {
		return fmt.Errorf("error deleting kvstore %s: %w", kvStoreName, err)
	}
	return nil
}

// Subscription operations.
func (c *cclient) CreateSubscription(ctx context.Context, projectID string, subscription *cloudruntime_client.DaprSubscription) error {
	if err := c.catalyst.CreateSubscription(ctx, projectID, subscription); err != nil {
		return fmt.Errorf("error creating subscription: %w", err)
	}
	return nil
}

func (c *cclient) GetSubscription(ctx context.Context, projectName string, subscriptionName string, qp *cloudruntime_client.DescribeDaprSubscriptionParams) (*cloudruntime_client.DaprSubscription, error) {
	subscription, err := c.catalyst.GetSubscription(ctx, projectName, subscriptionName, qp)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func (c *cclient) UpdateSubscription(ctx context.Context, projectName string, subscriptionName string, subscription *cloudruntime_client.DaprSubscription) error {
	if err := c.catalyst.PatchSubscription(ctx, projectName, subscriptionName, subscription); err != nil {
		return fmt.Errorf("error updating subscription %s: %w", subscriptionName, err)
	}
	return nil
}

func (c *cclient) DeleteSubscription(ctx context.Context, projectName string, subscriptionName string) error {
	if err := c.catalyst.DeleteSubscription(ctx, projectName, subscriptionName); err != nil {
		return fmt.Errorf("error deleting subscription %s: %w", subscriptionName, err)
	}
	return nil
}

// Resiliency operations.
func (c *cclient) CreateResiliency(ctx context.Context, projectID string, resiliency *cloudruntime_client.DaprResiliency) error {
	if err := c.catalyst.CreateResiliency(ctx, projectID, resiliency); err != nil {
		return fmt.Errorf("error creating resiliency: %w", err)
	}
	return nil
}

func (c *cclient) GetResiliency(ctx context.Context, projectName string, resiliencyName string, qp *cloudruntime_client.DescribeDaprResiliencyParams) (*cloudruntime_client.DaprResiliency, error) {
	resiliency, err := c.catalyst.GetResiliency(ctx, projectName, resiliencyName, qp)
	if err != nil {
		return nil, err
	}
	return resiliency, nil
}

func (c *cclient) UpdateResiliency(ctx context.Context, projectID string, resiliencyName string, resiliency *cloudruntime_client.DaprResiliency) error {
	if err := c.catalyst.PatchResiliency(ctx, projectID, resiliencyName, resiliency); err != nil {
		return fmt.Errorf("error updating resiliency %s: %w", resiliencyName, err)
	}
	return nil
}

func (c *cclient) DeleteResiliency(ctx context.Context, projectID string, resiliencyName string) error {
	if err := c.catalyst.DeleteResiliency(ctx, projectID, resiliencyName); err != nil {
		return fmt.Errorf("error deleting resiliency %s: %w", resiliencyName, err)
	}
	return nil
}

// Configuration operations.
func (c *cclient) CreateConfiguration(ctx context.Context, projectId string, config *cloudruntime_client.DaprConfiguration) error {
	if err := c.catalyst.CreateConfiguration(ctx, projectId, config); err != nil {
		return fmt.Errorf("error creating configuration: %w", err)
	}
	return nil
}

func (c *cclient) GetConfiguration(ctx context.Context, projectId string, configName string, qp *cloudruntime_client.DescribeDaprConfigurationParams) (*cloudruntime_client.DaprConfiguration, error) {
	config, err := c.catalyst.GetConfiguration(ctx, projectId, configName, qp)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c *cclient) UpdateConfiguration(ctx context.Context, projectId string, configName string, config *cloudruntime_client.DaprConfiguration) error {
	if err := c.catalyst.PatchConfiguration(ctx, projectId, configName, config); err != nil {
		return fmt.Errorf("error updating configuration %s: %w", configName, err)
	}
	return nil
}

func (c *cclient) DeleteConfiguration(ctx context.Context, projectId string, configName string) error {
	if err := c.catalyst.DeleteConfiguration(ctx, projectId, configName); err != nil {
		return fmt.Errorf("error deleting configuration %s: %w", configName, err)
	}
	return nil
}
