package pubsub

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading pubsub",
		map[string]interface{}{
			"project_name": m.GetProjectName(),
			"name":         m.GetName(),
		})

	pubsub, err := client.GetPubSub(ctx, m.GetProjectName(), m.GetName(), &cloudruntime_client.DescribePubSubParams{})
	if err != nil {
		return fmt.Errorf("error getting pubsub: %w", err)
	}

	tflog.Debug(ctx, "read pubsub",
		map[string]interface{}{
			"pubsub": pubsub,
		})

	if pubsub.Metadata != nil && pubsub.Metadata.Name != nil {
		m.SetName(*pubsub.Metadata.Name)
	}

	if pubsub.Spec != nil {
		if pubsub.Spec.ComponentName != nil {
			m.SetComponentName(*pubsub.Spec.ComponentName)
		}
		if pubsub.Spec.CreateComponent != nil {
			m.SetCreateComponent(*pubsub.Spec.CreateComponent)
		}

		// Handle scopes
		if pubsub.Spec.Scopes != nil && len(*pubsub.Spec.Scopes) > 0 {
			scopes := make([]attr.Value, 0, len(*pubsub.Spec.Scopes))
			for _, scope := range *pubsub.Spec.Scopes {
				scopes = append(scopes, types.StringValue(scope))
			}
			m.Scopes = types.ListValueMust(types.StringType, scopes)
		} else {
			m.Scopes = types.ListNull(types.StringType)
		}
	}

	// Set status (read-only)
	if pubsub.Status != nil {
		m.SetStatus(pubsub.GetStatus())
	} else {
		m.SetStatus("")
	}

	m.Log(ctx, "read pubsub model")

	return nil
}

func toAPIScopes(ctx context.Context, scopesList types.List) *cloudruntime_client.DaprScopes {
	if scopesList.IsNull() || scopesList.IsUnknown() {
		return nil
	}

	var scopes []string
	diags := scopesList.ElementsAs(ctx, &scopes, false)
	if diags.HasError() {
		return nil
	}

	result := cloudruntime_client.DaprScopes(scopes)
	return &result
}
