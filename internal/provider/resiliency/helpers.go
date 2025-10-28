package resiliency

import (
	"context"
	"encoding/json"
	"fmt"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v3"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

// YAML-friendly types for proper marshaling/unmarshaling.
type resiliencySpecForYAML struct {
	Policies *policiesForYAML `yaml:"policies,omitempty"`
	Targets  *targetsForYAML  `yaml:"targets,omitempty"`
}

type policiesForYAML struct {
	Timeouts        map[string]interface{} `yaml:"timeouts,omitempty"`
	Retries         map[string]interface{} `yaml:"retries,omitempty"`
	CircuitBreakers map[string]interface{} `yaml:"circuitBreakers,omitempty"`
}

type targetsForYAML struct {
	Apps       map[string]interface{} `yaml:"apps,omitempty"`
	Actors     map[string]interface{} `yaml:"actors,omitempty"`
	Components map[string]interface{} `yaml:"components,omitempty"`
}

// toAPIScopes converts a Terraform List to API DaprScopes.
func toAPIScopes(ctx context.Context, scopesList types.List) *cloudruntime_client.DaprScopes {
	if scopesList.IsNull() || scopesList.IsUnknown() {
		return nil
	}

	var scopes []string
	if diags := scopesList.ElementsAs(ctx, &scopes, false); diags.HasError() {
		tflog.Error(ctx, "error converting scopes to slice", map[string]interface{}{
			"diagnostics": diags,
		})
		return nil
	}

	return (*cloudruntime_client.DaprScopes)(&scopes)
}

// toAPISpec converts YAML string to DaprResiliencySpec.
func toAPISpec(ctx context.Context, specYAML types.String) (*cloudruntime_client.DaprResiliencySpec, error) {
	if specYAML.IsNull() || specYAML.IsUnknown() {
		return &cloudruntime_client.DaprResiliencySpec{}, nil
	}

	// Unmarshal YAML into intermediate type
	var yamlSpec resiliencySpecForYAML
	if err := yaml.Unmarshal([]byte(specYAML.ValueString()), &yamlSpec); err != nil {
		tflog.Error(ctx, "error unmarshaling resiliency spec YAML", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Convert to JSON then unmarshal into API type (which has JSON tags)
	jsonBytes, err := json.Marshal(yamlSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	var apiSpec cloudruntime_client.DaprResiliencySpec
	if err := json.Unmarshal(jsonBytes, &apiSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into API type: %w", err)
	}

	return &apiSpec, nil
}

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading resiliency",
		map[string]interface{}{
			"project_id": m.GetProjectID(),
			"name":       m.GetName(),
		})

	resiliency, err := client.GetResiliency(ctx, m.GetProjectID(), m.GetName(), &cloudruntime_client.DescribeDaprResiliencyParams{})
	if err != nil {
		return fmt.Errorf("error getting resiliency: %w", err)
	}

	tflog.Debug(ctx, "read resiliency",
		map[string]interface{}{
			"resiliency": resiliency,
		})

	if resiliency.Metadata != nil && resiliency.Metadata.Name != nil {
		m.SetName(*resiliency.Metadata.Name)
	}

	// Set scopes
	if resiliency.Scopes != nil && len(*resiliency.Scopes) > 0 {
		scopesList, diags := types.ListValueFrom(ctx, types.StringType, *resiliency.Scopes)
		if diags.HasError() {
			tflog.Error(ctx, "error converting scopes to list", map[string]interface{}{
				"diagnostics": diags,
			})
		} else {
			m.Scopes = scopesList
		}
	} else {
		m.Scopes = types.ListNull(types.StringType)
	}

	// Set status
	if resiliency.Status != nil {
		m.SetStatus(resiliency.GetStatus())
	} else {
		m.Status = types.StringNull()
	}

	// Note: We intentionally DO NOT update the spec field here.
	// The spec is set from the plan during Create/Update and preserved in state.
	// This avoids YAML formatting differences between user input and API responses.

	m.Log(ctx, "read resiliency model")

	return nil
}
