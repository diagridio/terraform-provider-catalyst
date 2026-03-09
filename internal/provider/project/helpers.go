package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	catalyst_client "github.com/diagridio/cloudgrid/sdk/go/pkg/catalyst/client"
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

	project, err := client.GetProject(ctx, m.GetName(), &catalyst_client.DescribeProjectParams{})
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

	m.SetDefaultAgentInfrastructureEnabled(project.Spec.DefaultAgentInfrastructureEnabled)
	m.SetDefaultKVStoreEnabled(project.Spec.DefaultKVStoreEnabled)
	m.SetDefaultPubsubEnabled(project.Spec.DefaultPubsubEnabled)
	m.SetDefaultWorkflowStoreEnabled(project.Spec.DefaultWorkflowStoreEnabled)
	m.SetDisableAppTunnels(project.Spec.DisableAppTunnels)
	m.SetPrivateRegion(project.Spec.PrivateRegion)
	if project.Spec.GlobalAppId != nil {
		m.SetGlobalAppIdMaxBodySize(project.Spec.GlobalAppId.MaxBodySize)
	} else {
		m.SetGlobalAppIdMaxBodySize(nil)
	}

	return nil

}
