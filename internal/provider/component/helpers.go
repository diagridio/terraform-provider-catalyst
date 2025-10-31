package component

import (
	"context"
	"fmt"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v3"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

// toAPIScopes converts Terraform list to API scopes slice.
func toAPIScopes(ctx context.Context, scopesList types.List) cloudruntime_client.DaprScopes {
	if scopesList.IsNull() || scopesList.IsUnknown() {
		return nil
	}

	var scopes []string
	scopesList.ElementsAs(ctx, &scopes, false)
	return scopes
}

// metadataItemForYAML is a helper type for proper YAML marshaling/unmarshaling.
type metadataItemForYAML struct {
	Name         string  `yaml:"name"`
	Value        *string `yaml:"value,omitempty"`
	SecretKeyRef *struct {
		Name string `yaml:"name"`
		Key  string `yaml:"key"`
	} `yaml:"secretKeyRef,omitempty"`
}

// toAPISpec converts YAML string to metadata array for Component spec.
func toAPISpec(_ context.Context, specString types.String) (*[]cloudruntime_client.DaprMetadataItem, error) {
	if specString.IsNull() || specString.IsUnknown() {
		return nil, nil
	}

	// First unmarshal into YAML-friendly type
	var yamlItems []metadataItemForYAML
	if err := yaml.Unmarshal([]byte(specString.ValueString()), &yamlItems); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec YAML: %w", err)
	}

	// Convert to API type
	metadata := make([]cloudruntime_client.DaprMetadataItem, len(yamlItems))
	for i, yamlItem := range yamlItems {
		item := cloudruntime_client.DaprMetadataItem{
			Name: &yamlItem.Name,
		}
		if yamlItem.Value != nil {
			var val interface{} = *yamlItem.Value
			item.Value = &val
		}
		if yamlItem.SecretKeyRef != nil {
			item.SecretKeyRef = &struct {
				Key  *string `json:"key,omitempty"`
				Name *string `json:"name,omitempty"`
			}{
				Name: &yamlItem.SecretKeyRef.Name,
				Key:  &yamlItem.SecretKeyRef.Key,
			}
		}
		metadata[i] = item
	}

	return &metadata, nil
}

// normalizeYAML normalizes a YAML string by unmarshaling and remarshaling it.
// This ensures consistent formatting regardless of input formatting.
func normalizeYAML(yamlStr string) (string, error) {
	var items []metadataItemForYAML
	if err := yaml.Unmarshal([]byte(yamlStr), &items); err != nil {
		return "", fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	normalized, err := yaml.Marshal(items)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(normalized), nil
}

// toYAML converts API metadata items to YAML representation with omitempty.
func metadataToYAML(metadata *[]cloudruntime_client.DaprMetadataItem) (string, error) {
	if metadata == nil || len(*metadata) == 0 {
		return "", nil
	}

	// Convert to helper type that properly handles omitempty
	yamlItems := make([]metadataItemForYAML, len(*metadata))
	for i, item := range *metadata {
		yamlItem := metadataItemForYAML{
			Name: *item.Name,
		}
		if item.Value != nil {
			// Convert interface{} to string
			if strVal, ok := (*item.Value).(string); ok {
				yamlItem.Value = &strVal
			}
		}
		if item.SecretKeyRef != nil {
			yamlItem.SecretKeyRef = &struct {
				Name string `yaml:"name"`
				Key  string `yaml:"key"`
			}{
				Name: *item.SecretKeyRef.Name,
				Key:  *item.SecretKeyRef.Key,
			}
		}
		yamlItems[i] = yamlItem
	}

	specYAML, err := yaml.Marshal(yamlItems)
	if err != nil {
		return "", fmt.Errorf("failed to marshal spec metadata to YAML: %w", err)
	}
	return string(specYAML), nil
}

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading component",
		map[string]interface{}{
			"project_name": m.GetProjectName(),
			"name":         m.GetName(),
		})

	showSensitive := true
	component, err := client.GetComponent(ctx, m.GetProjectName(), m.GetName(), &cloudruntime_client.DescribeDaprComponentParams{
		Showsensitive: &showSensitive,
	})
	if err != nil {
		return fmt.Errorf("error getting component: %w", err)
	}

	tflog.Debug(ctx, "read component",
		map[string]interface{}{
			"component": component,
		})

	if component.Metadata != nil && component.Metadata.Name != nil {
		m.SetName(*component.Metadata.Name)
	}

	if component.Spec != nil {
		if component.Spec.Type != nil {
			m.SetType(*component.Spec.Type)
		}
		if component.Spec.Version != nil {
			m.SetVersion(*component.Spec.Version)
		}

		// Convert metadata to YAML for Terraform state
		if component.Spec.Metadata != nil && len(*component.Spec.Metadata) > 0 {
			specYAML, err := metadataToYAML(component.Spec.Metadata)
			if err != nil {
				return err
			}
			if specYAML != "" {
				// Normalize the YAML to ensure consistent formatting
				normalized, err := normalizeYAML(specYAML)
				if err != nil {
					return fmt.Errorf("error normalizing spec YAML: %w", err)
				}
				m.SetSpec(normalized)
			}
		} else {
			m.Spec = types.StringNull()
		}
	}

	// Handle auth (secret_store)
	if component.Auth != nil && component.Auth.SecretStore != nil {
		m.SetSecretStore(*component.Auth.SecretStore)
	} else {
		m.SetSecretStore("")
	}

	// Handle scopes
	if component.Scopes != nil && len(*component.Scopes) > 0 {
		scopesList, diags := types.ListValueFrom(ctx, types.StringType, *component.Scopes)
		if diags.HasError() {
			return fmt.Errorf("failed to convert scopes: %s", diags.Errors())
		}
		m.Scopes = scopesList
	} else {
		m.Scopes = types.ListNull(types.StringType)
	}

	// Set status (read-only)
	if component.Status != nil {
		m.SetStatus(component.GetStatus())
	} else {
		m.SetStatus("")
	}

	m.Log(ctx, "read component model")

	return nil
}
