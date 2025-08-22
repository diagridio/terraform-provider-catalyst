package catalyst

import (
	"context"
	"fmt"
	"net/http"

	"github.com/diagridio/diagrid-cloud-go/cloudruntime"
	"github.com/diagridio/diagrid-cloud-go/management"
	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	conductor_client "github.com/diagridio/diagrid-cloud-go/pkg/conductor/client"
)

type Client interface {
	GetUserOrg(context.Context) (*conductor_client.Organization, error)

	CreateRegion(ctx context.Context, region *cloudruntime_client.Region) (string, error)
	GetRegion(ctx context.Context, name string) (*cloudruntime_client.Region, error)
	UpdateRegion(ctx context.Context, region *cloudruntime_client.Region) error
	DeleteRegion(ctx context.Context, name string) error

	GetProject(ctx context.Context, id string, qp *cloudruntime_client.DescribeProjectParams) (*cloudruntime_client.Project, error)
	CreateProject(ctx context.Context, project *cloudruntime_client.Project) error
	UpdateProject(ctx context.Context, prj *cloudruntime_client.Project) error
	DeleteProject(ctx context.Context, id string) error
}

type cclient struct {
	management *management.ManagementClient
	catalyst   cloudruntime.CloudruntimeAPIClient
}

var (
	ErrAPIKeyNotFound   = fmt.Errorf("API key not found in environment variable CATALYST_API_KEY or provider configuration block api_key attribute")
	ErrEndpointNotFound = fmt.Errorf("endpoint not found in environment variable CATALYST_API_ENDPOINT or provider configuration block endpoint attribute")
)

func NewClient(endpoint, apiKey string) (Client, error) {
	if apiKey == "" {
		return nil, ErrAPIKeyNotFound
	}
	if endpoint == "" {
		return nil, ErrEndpointNotFound
	}

	// Example client configuration for data sources and resources
	maxRetries := 1
	mc, err := management.NewManagementClientWithExponentialBackoff(http.DefaultClient,
		endpoint,
		maxRetries,
		management.WithAPIKeyToken(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating management client: %w", err)
	}

	catalystClient, err := cloudruntime.NewCloudruntimeClientWithExponentialBackoff(http.DefaultClient,
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
	if err := c.catalyst.CreateProject(ctx, project); err != nil {
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
