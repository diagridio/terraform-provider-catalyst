package catalyst

import (
	"context"
	"fmt"
	"net/http"

	"github.com/diagridio/diagrid-cloud-go/cloudruntime"
	"github.com/diagridio/diagrid-cloud-go/management"
	diagrid_client "github.com/diagridio/diagrid-cloud-go/pkg/client"
	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
)

type Client interface {
	GetUserOrg(context.Context) (*diagrid_client.Organization, error)
	ListRegions(context.Context) (*[]cloudruntime_client.Region, error)
	GetProject(ctx context.Context, id string, qp *cloudruntime_client.DescribeProjectParams) (*cloudruntime_client.Project, error)
	CreateProject(ctx context.Context, project *cloudruntime_client.Project) error
	PatchProject(ctx context.Context, prj *cloudruntime_client.Project) error
	DeleteProject(ctx context.Context, id string) error
}

type cclient struct {
	management *management.ManagementClient
	catalyst   cloudruntime.CloudruntimeAPIClient
}

func NewClient(endpoint, apiKey string) (Client, error) {
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

func (c *cclient) GetUserOrg(ctx context.Context) (*diagrid_client.Organization, error) {
	org, err := c.management.GetUserOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user org: %w", err)
	}

	return org, nil
}

func (c *cclient) ListRegions(ctx context.Context) (*[]cloudruntime_client.Region, error) {
	regions, err := c.catalyst.ListRegions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing catalyst regions: %w", err)
	}

	return regions, nil
}

func (c *cclient) GetProject(ctx context.Context, id string, qp *cloudruntime_client.DescribeProjectParams) (*cloudruntime_client.Project, error) {
	project, err := c.catalyst.GetProject(ctx, id, qp)
	if err != nil {
		return nil, fmt.Errorf("error getting project %s: %w", id, err)
	}

	return project, nil
}

func (c *cclient) CreateProject(ctx context.Context, project *cloudruntime_client.Project) error {
	if err := c.catalyst.CreateProject(ctx, project); err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	return nil
}

func (c *cclient) PatchProject(ctx context.Context, project *cloudruntime_client.Project) error {
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
