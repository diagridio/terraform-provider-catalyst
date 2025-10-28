package subscription

import (
	"context"
	"fmt"

	cloudruntime_client "github.com/diagridio/diagrid-cloud-go/pkg/cloudruntime/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v3"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

// YAML-friendly types for proper marshaling/unmarshaling.
type subscriptionSpecForYAML struct {
	Routes          *routesForYAML        `yaml:"routes,omitempty"`
	BulkSubscribe   *bulkSubscribeForYAML `yaml:"bulkSubscribe,omitempty"`
	DeadLetterTopic *string               `yaml:"deadLetterTopic,omitempty"`
	Metadata        map[string]string     `yaml:"metadata,omitempty"`
	Dynamic         *bool                 `yaml:"dynamic,omitempty"`
}

type routesForYAML struct {
	Rules   []ruleForYAML `yaml:"rules,omitempty"`
	Default *string       `yaml:"default,omitempty"`
}

type ruleForYAML struct {
	Match *string `yaml:"match,omitempty"`
	Path  *string `yaml:"path,omitempty"`
}

type bulkSubscribeForYAML struct {
	Enabled            *bool `yaml:"enabled,omitempty"`
	MaxMessagesCount   *int  `yaml:"maxMessagesCount,omitempty"`
	MaxAwaitDurationMs *int  `yaml:"maxAwaitDurationMs,omitempty"`
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

// toAPISpec converts YAML string to DaprSubscriptionSpec.
func toAPISpec(_ context.Context, specString types.String) (*cloudruntime_client.DaprSubscriptionSpec, error) {
	if specString.IsNull() || specString.IsUnknown() {
		return &cloudruntime_client.DaprSubscriptionSpec{}, nil
	}

	// Unmarshal into YAML-friendly type
	var yamlSpec subscriptionSpecForYAML
	if err := yaml.Unmarshal([]byte(specString.ValueString()), &yamlSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec YAML: %w", err)
	}

	// Convert to API type
	apiSpec := &cloudruntime_client.DaprSubscriptionSpec{}

	if yamlSpec.Routes != nil {
		routes := &cloudruntime_client.SubscriptionSpecRoutes{}
		if yamlSpec.Routes.Default != nil {
			routes.Default = yamlSpec.Routes.Default
		}
		if len(yamlSpec.Routes.Rules) > 0 {
			rules := make([]cloudruntime_client.SubscriptionRule, len(yamlSpec.Routes.Rules))
			for i, rule := range yamlSpec.Routes.Rules {
				rules[i] = cloudruntime_client.SubscriptionRule{
					Match: rule.Match,
					Path:  rule.Path,
				}
			}
			routes.Rules = &rules
		}
		apiSpec.Routes = routes
	}

	if yamlSpec.BulkSubscribe != nil {
		bulkSub := &cloudruntime_client.DaprSubscriptionSpecBulkSubscribe{}
		if yamlSpec.BulkSubscribe.Enabled != nil {
			bulkSub.Enabled = yamlSpec.BulkSubscribe.Enabled
		}
		if yamlSpec.BulkSubscribe.MaxMessagesCount != nil {
			bulkSub.MaxMessagesCount = yamlSpec.BulkSubscribe.MaxMessagesCount
		}
		if yamlSpec.BulkSubscribe.MaxAwaitDurationMs != nil {
			bulkSub.MaxAwaitDurationMs = yamlSpec.BulkSubscribe.MaxAwaitDurationMs
		}
		apiSpec.BulkSubscribe = bulkSub
	}

	if yamlSpec.DeadLetterTopic != nil {
		apiSpec.DeadLetterTopic = yamlSpec.DeadLetterTopic
	}

	if len(yamlSpec.Metadata) > 0 {
		metadata := &cloudruntime_client.DaprSubscriptionSpec_Metadata{
			AdditionalProperties: yamlSpec.Metadata,
		}
		apiSpec.Metadata = metadata
	}

	if yamlSpec.Dynamic != nil {
		apiSpec.Dynamic = yamlSpec.Dynamic
	}

	return apiSpec, nil
}

// specToYAML converts API subscription spec to YAML string.
func specToYAML(spec *cloudruntime_client.DaprSubscriptionSpec) (string, error) {
	if spec == nil {
		return "", nil
	}

	yamlSpec := subscriptionSpecForYAML{}

	// Convert routes
	if spec.Routes != nil {
		routes := &routesForYAML{}
		if spec.Routes.Rules != nil && len(*spec.Routes.Rules) > 0 {
			rules := make([]ruleForYAML, len(*spec.Routes.Rules))
			for i, rule := range *spec.Routes.Rules {
				rules[i] = ruleForYAML{
					Match: rule.Match,
					Path:  rule.Path,
				}
			}
			routes.Rules = rules
		}
		if spec.Routes.Default != nil {
			routes.Default = spec.Routes.Default
		}
		if len(routes.Rules) > 0 || routes.Default != nil {
			yamlSpec.Routes = routes
		}
	}

	// Convert bulk subscribe
	if spec.BulkSubscribe != nil {
		bulkSub := &bulkSubscribeForYAML{}
		if spec.BulkSubscribe.Enabled != nil {
			bulkSub.Enabled = spec.BulkSubscribe.Enabled
		}
		if spec.BulkSubscribe.MaxMessagesCount != nil {
			bulkSub.MaxMessagesCount = spec.BulkSubscribe.MaxMessagesCount
		}
		if spec.BulkSubscribe.MaxAwaitDurationMs != nil {
			bulkSub.MaxAwaitDurationMs = spec.BulkSubscribe.MaxAwaitDurationMs
		}
		if bulkSub.Enabled != nil || bulkSub.MaxMessagesCount != nil || bulkSub.MaxAwaitDurationMs != nil {
			yamlSpec.BulkSubscribe = bulkSub
		}
	}

	// Convert metadata - preserve order from input if possible
	if spec.Metadata != nil && len(spec.Metadata.AdditionalProperties) > 0 {
		yamlSpec.Metadata = spec.Metadata.AdditionalProperties
	}

	// Convert other fields
	if spec.DeadLetterTopic != nil {
		yamlSpec.DeadLetterTopic = spec.DeadLetterTopic
	}
	if spec.Dynamic != nil {
		yamlSpec.Dynamic = spec.Dynamic
	}

	// If everything is empty, return empty string
	if yamlSpec.Routes == nil && yamlSpec.BulkSubscribe == nil && yamlSpec.DeadLetterTopic == nil &&
		len(yamlSpec.Metadata) == 0 && yamlSpec.Dynamic == nil {
		return "", nil
	}

	// Marshal to YAML with 4-space indentation to match input format
	yamlBytes, err := yaml.Marshal(yamlSpec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal spec to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading subscription",
		map[string]interface{}{
			"project_name": m.GetProjectName(),
			"name":         m.GetName(),
		})

	subscription, err := client.GetSubscription(ctx, m.GetProjectName(), m.GetName(), &cloudruntime_client.DescribeDaprSubscriptionParams{})
	if err != nil {
		return fmt.Errorf("error getting subscription: %w", err)
	}

	tflog.Debug(ctx, "read subscription",
		map[string]interface{}{
			"subscription": subscription,
		})

	if subscription.Metadata != nil && subscription.Metadata.Name != nil {
		m.SetName(*subscription.Metadata.Name)
	}

	if subscription.Spec != nil {
		if subscription.Spec.Pubsubname != nil {
			m.SetPubsubName(*subscription.Spec.Pubsubname)
		}
		if subscription.Spec.Topic != nil {
			m.SetTopic(*subscription.Spec.Topic)
		}

		// Convert spec fields (routes, bulk_subscribe, metadata, etc.) to YAML
		// We only marshal the parts that aren't basic required fields (pubsubname/topic)
		if subscription.Spec.Routes != nil || subscription.Spec.DeadLetterTopic != nil ||
			subscription.Spec.BulkSubscribe != nil || subscription.Spec.Metadata != nil ||
			subscription.Spec.Dynamic != nil {

			specYAML, err := specToYAML(subscription.Spec)
			if err != nil {
				return err
			}
			if specYAML != "" {
				m.SetSpec(specYAML)
			}
		} else {
			m.Spec = types.StringNull()
		}
	}

	// Set scopes (at DaprSubscription level, not Spec)
	if subscription.Scopes != nil && len(*subscription.Scopes) > 0 {
		scopesList, diags := types.ListValueFrom(ctx, types.StringType, *subscription.Scopes)
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
	if subscription.Status != nil {
		m.SetStatus(subscription.GetStatus())
	} else {
		m.Status = types.StringNull()
	}

	m.Log(ctx, "read subscription model")

	return nil
}
