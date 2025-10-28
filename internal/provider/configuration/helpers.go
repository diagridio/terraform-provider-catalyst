package configuration

import (
	"context"
	"fmt"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v3"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

// toAPISpec converts YAML string to DaprConfigurationSpec.
func toAPISpec(_ context.Context, specString types.String) (*cloudruntime_client.DaprConfigurationSpec, error) {
	if specString.IsNull() || specString.IsUnknown() {
		return &cloudruntime_client.DaprConfigurationSpec{}, nil
	}

	var spec cloudruntime_client.DaprConfigurationSpec
	if err := yaml.Unmarshal([]byte(specString.ValueString()), &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec YAML: %w", err)
	}

	return &spec, nil
}

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading configuration",
		map[string]interface{}{
			"project_id": m.GetProjectID(),
			"name":       m.GetName(),
		})

	configuration, err := client.GetConfiguration(ctx, m.GetProjectID(), m.GetName(), &cloudruntime_client.DescribeDaprConfigurationParams{})
	if err != nil {
		return fmt.Errorf("error getting configuration: %w", err)
	}

	tflog.Debug(ctx, "read configuration",
		map[string]interface{}{
			"configuration": configuration,
		})

	if configuration.Metadata != nil && configuration.Metadata.Name != nil {
		m.SetName(*configuration.Metadata.Name)
	}

	// Note: We intentionally DO NOT update the spec field here.
	// The spec is set from the plan during Create/Update and preserved in state.
	// This avoids YAML formatting differences between user input and API responses.

	// Set status
	if configuration.Status != nil && configuration.Status.Status != nil {
		m.SetStatus(*configuration.Status.Status)
	} else {
		m.Status = types.StringNull()
	}

	m.Log(ctx, "read configuration model")

	return nil
}
