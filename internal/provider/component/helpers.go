package component

import (
	"context"
	"fmt"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

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

func expandComponentSpec(spec *specModel) (*cloudruntime_client.DaprComponentSpec, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec must be provided")
	}

	if spec.Type.IsNull() || spec.Type.IsUnknown() {
		return nil, fmt.Errorf("spec.type must be provided")
	}

	typ := spec.Type.ValueString()
	apiSpec := &cloudruntime_client.DaprComponentSpec{
		Type: &typ,
	}

	if !spec.Version.IsNull() && !spec.Version.IsUnknown() {
		version := spec.Version.ValueString()
		apiSpec.Version = &version
	}

	if len(spec.Metadata) > 0 {
		metadata := make([]cloudruntime_client.DaprMetadataItem, len(spec.Metadata))
		for i, item := range spec.Metadata {
			if item.Name.IsNull() || item.Name.IsUnknown() || item.Name.ValueString() == "" {
				return nil, fmt.Errorf("spec.metadata[%d].name must be provided", i)
			}

			name := item.Name.ValueString()
			metadata[i].Name = &name

			if !item.Value.IsNull() && !item.Value.IsUnknown() {
				value := item.Value.ValueString()
				var iface interface{} = value
				metadata[i].Value = &iface
			}

			if item.SecretKeyRef != nil {
				if item.SecretKeyRef.Name.IsNull() || item.SecretKeyRef.Name.IsUnknown() || item.SecretKeyRef.Name.ValueString() == "" {
					return nil, fmt.Errorf("spec.metadata[%d].secret_key_ref.name must be provided", i)
				}
				if item.SecretKeyRef.Key.IsNull() || item.SecretKeyRef.Key.IsUnknown() || item.SecretKeyRef.Key.ValueString() == "" {
					return nil, fmt.Errorf("spec.metadata[%d].secret_key_ref.key must be provided", i)
				}

				keyName := item.SecretKeyRef.Name.ValueString()
				keyKey := item.SecretKeyRef.Key.ValueString()
				metadata[i].SecretKeyRef = &struct {
					Key  *string `json:"key,omitempty"`
					Name *string `json:"name,omitempty"`
				}{
					Name: &keyName,
					Key:  &keyKey,
				}
			}
		}
		apiSpec.Metadata = &metadata
	}

	return apiSpec, nil
}

func flattenMetadataItems(items *[]cloudruntime_client.DaprMetadataItem) []metadataItemModel {
	if items == nil {
		return nil
	}

	if len(*items) == 0 {
		return []metadataItemModel{}
	}

	metadata := make([]metadataItemModel, len(*items))
	for i, item := range *items {
		metadata[i].Name = types.StringNull()
		metadata[i].Value = types.StringNull()

		if item.Name != nil && *item.Name != "" {
			metadata[i].Name = types.StringValue(*item.Name)
		}

		if item.Value != nil {
			metadata[i].Value = types.StringValue(fmt.Sprint(*item.Value))
		}

		if item.SecretKeyRef != nil && item.SecretKeyRef.Name != nil && item.SecretKeyRef.Key != nil {
			if *item.SecretKeyRef.Name != "" && *item.SecretKeyRef.Key != "" {
				metadata[i].SecretKeyRef = &secretKeyRefModel{
					Name: types.StringValue(*item.SecretKeyRef.Name),
					Key:  types.StringValue(*item.SecretKeyRef.Key),
				}
			}
		}
	}

	return metadata
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

	spec := m.ensureSpec()
	if component.Spec != nil {
		if component.Spec.Type != nil && *component.Spec.Type != "" {
			m.SetType(*component.Spec.Type)
		} else {
			spec.Type = types.StringNull()
		}
		if component.Spec.Version != nil && *component.Spec.Version != "" {
			m.SetVersion(*component.Spec.Version)
		} else {
			spec.Version = types.StringNull()
		}
		spec.Metadata = flattenMetadataItems(component.Spec.Metadata)
	} else {
		spec.Type = types.StringNull()
		spec.Version = types.StringNull()
		spec.Metadata = nil
	}

	if component.Auth != nil && component.Auth.SecretStore != nil && *component.Auth.SecretStore != "" {
		m.Auth = &authModel{SecretStore: types.StringValue(*component.Auth.SecretStore)}
	} else {
		m.Auth = nil
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
