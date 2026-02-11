package kvstore

import (
	"context"
	"fmt"

	catalyst_client "github.com/diagridio/cloudgrid/sdk/go/pkg/catalyst/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

func toAPIScopes(ctx context.Context, scopes types.List) *catalyst_client.DaprScopes {
	if scopes.IsNull() || scopes.IsUnknown() {
		return nil
	}

	var scopesSlice []string
	if diags := scopes.ElementsAs(ctx, &scopesSlice, false); diags.HasError() {
		tflog.Error(ctx, "error converting scopes to slice", map[string]interface{}{
			"diagnostics": diags,
		})
		return nil
	}

	return &scopesSlice
}

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading kvstore",
		map[string]interface{}{
			"project_name": m.GetProjectName(),
			"name":         m.GetName(),
		})

	kvstore, err := client.GetKVStore(ctx, m.GetProjectName(), m.GetName(), &catalyst_client.DescribeKVStoreParams{})
	if err != nil {
		return fmt.Errorf("error getting kvstore: %w", err)
	}

	tflog.Debug(ctx, "read kvstore",
		map[string]interface{}{
			"kvstore": kvstore,
		})

	if kvstore.Metadata != nil && kvstore.Metadata.Name != nil {
		m.SetName(*kvstore.Metadata.Name)
	}

	if kvstore.Spec != nil {
		if kvstore.Spec.ComponentName != nil {
			m.SetComponentName(*kvstore.Spec.ComponentName)
		}
		if kvstore.Spec.CreateComponent != nil {
			m.SetCreateComponent(*kvstore.Spec.CreateComponent)
		}

		// Set scopes
		if kvstore.Spec.Scopes != nil && len(*kvstore.Spec.Scopes) > 0 {
			scopesList, diags := types.ListValueFrom(ctx, types.StringType, *kvstore.Spec.Scopes)
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
	}

	// Set status
	if kvstore.Status != nil && kvstore.Status.Status != nil {
		m.SetStatus(*kvstore.Status.Status)
	}

	m.Log(ctx, "read kvstore model")

	return nil
}
