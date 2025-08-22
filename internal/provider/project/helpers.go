package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading project",
		map[string]interface{}{
			"name": m.GetName(),
		})

	project, err := client.GetProject(ctx, m.GetName(), &cloudruntime_client.DescribeProjectParams{})
	if err != nil {
		return fmt.Errorf("error getting project: %w", err)
	}

	m.Log(ctx, "read project")

	m.SetName(*project.Metadata.Name)
	m.SetRegion(*project.Spec.Region)
	if project.Status.Endpoints != nil &&
		project.Status.Endpoints.Http != nil &&
		project.Status.Endpoints.Http.Url != nil {
		m.SetHTTPEndpoint(*project.Status.Endpoints.Http.Url)
	}
	if project.Status.Endpoints != nil &&
		project.Status.Endpoints.Grpc != nil &&
		project.Status.Endpoints.Grpc.Url != nil {
		m.SetGRPCEndpoint(*project.Status.Endpoints.Grpc.Url)
	}

	return nil

}
