package subscription

import (
	"context"
	"fmt"

	catalyst_client "github.com/diagridio/cloudgrid/sdk/go/pkg/catalyst/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

// toAPIScopes converts a Terraform List to API DaprScopes.
func toAPIScopes(ctx context.Context, scopesList types.List) *catalyst_client.DaprScopes {
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

	return &scopes
}

func expandSubscriptionSpec(ctx context.Context, model *specModel) (*catalyst_client.DaprSubscriptionSpec, error) {
	if model == nil {
		return nil, nil
	}

	spec := &catalyst_client.DaprSubscriptionSpec{}
	hasData := false

	if routes := expandRoutes(model.Routes); routes != nil {
		spec.Routes = routes
		hasData = true
	}

	if bulk := expandBulkSubscribe(model.BulkSubscribe); bulk != nil {
		spec.BulkSubscribe = bulk
		hasData = true
	}

	if !model.DeadLetterTopic.IsNull() && !model.DeadLetterTopic.IsUnknown() {
		if value := model.DeadLetterTopic.ValueString(); value != "" {
			valueCopy := value
			spec.DeadLetterTopic = &valueCopy
			hasData = true
		}
	}

	if !model.Metadata.IsNull() && !model.Metadata.IsUnknown() {
		metadata, err := mapToStringMap(ctx, model.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to expand metadata: %w", err)
		}
		if len(metadata) > 0 {
			spec.Metadata = &metadata
			hasData = true
		}
	}

	if !hasData {
		return nil, nil
	}

	return spec, nil
}

func expandRoutes(model *routesModel) *catalyst_client.SubscriptionSpecRoutes {
	if model == nil {
		return nil
	}

	routes := &catalyst_client.SubscriptionSpecRoutes{}
	hasData := false

	if !model.Default.IsNull() && !model.Default.IsUnknown() {
		if value := model.Default.ValueString(); value != "" {
			valueCopy := value
			routes.Default = &valueCopy
			hasData = true
		}
	}

	if len(model.Rules) > 0 {
		rules := make([]catalyst_client.SubscriptionRule, 0, len(model.Rules))
		for _, ruleModel := range model.Rules {
			rule := catalyst_client.SubscriptionRule{}
			ruleHasData := false

			if !ruleModel.Match.IsNull() && !ruleModel.Match.IsUnknown() {
				if value := ruleModel.Match.ValueString(); value != "" {
					rule.Match = &value
					ruleHasData = true
				}
			}

			if !ruleModel.Path.IsNull() && !ruleModel.Path.IsUnknown() {
				if value := ruleModel.Path.ValueString(); value != "" {
					rule.Path = &value
					ruleHasData = true
				}
			}

			if ruleHasData {
				rules = append(rules, rule)
			}
		}
		if len(rules) > 0 {
			routes.Rules = &rules
			hasData = true
		}
	}

	if !hasData {
		return nil
	}

	return routes
}

func expandBulkSubscribe(model *bulkSubscribeModel) *catalyst_client.DaprSubscriptionSpecBulkSubscribe {
	if model == nil {
		return nil
	}

	bulk := &catalyst_client.DaprSubscriptionSpecBulkSubscribe{}
	hasData := false

	if !model.Enabled.IsNull() && !model.Enabled.IsUnknown() {
		value := model.Enabled.ValueBool()
		bulk.Enabled = &value
		hasData = true
	}

	if !model.MaxMessagesCount.IsNull() && !model.MaxMessagesCount.IsUnknown() {
		value := int(model.MaxMessagesCount.ValueInt64())
		bulk.MaxMessagesCount = &value
		hasData = true
	}

	if !model.MaxAwaitDurationMs.IsNull() && !model.MaxAwaitDurationMs.IsUnknown() {
		value := int(model.MaxAwaitDurationMs.ValueInt64())
		bulk.MaxAwaitDurationMs = &value
		hasData = true
	}

	if !hasData {
		return nil
	}

	return bulk
}

func flattenSubscriptionSpec(ctx context.Context, api *catalyst_client.DaprSubscriptionSpec) (*specModel, error) {
	spec := &specModel{
		DeadLetterTopic: types.StringNull(),
		Metadata:        types.MapNull(types.StringType),
	}

	if api == nil {
		return spec, nil
	}

	if api.Routes != nil {
		spec.Routes = flattenRoutes(api.Routes)
	}

	if api.BulkSubscribe != nil {
		spec.BulkSubscribe = flattenBulkSubscribe(api.BulkSubscribe)
	}

	if api.DeadLetterTopic != nil && *api.DeadLetterTopic != "" {
		spec.DeadLetterTopic = types.StringValue(*api.DeadLetterTopic)
	}

	if api.Metadata != nil && len(*api.Metadata) > 0 {
		metadataValue, diags := types.MapValueFrom(ctx, types.StringType, *api.Metadata)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to flatten metadata: %v", diags.Errors())
		}
		spec.Metadata = metadataValue
	}

	return spec, nil
}

func flattenRoutes(api *catalyst_client.SubscriptionSpecRoutes) *routesModel {
	if api == nil {
		return nil
	}

	routes := &routesModel{
		Default: types.StringNull(),
	}
	hasData := false

	if api.Default != nil && *api.Default != "" {
		routes.Default = types.StringValue(*api.Default)
		hasData = true
	}

	if api.Rules != nil && len(*api.Rules) > 0 {
		rules := make([]routeRuleModel, 0, len(*api.Rules))
		for _, rule := range *api.Rules {
			model := routeRuleModel{
				Match: types.StringNull(),
				Path:  types.StringNull(),
			}
			ruleHasData := false

			if rule.Match != nil && *rule.Match != "" {
				model.Match = types.StringValue(*rule.Match)
				ruleHasData = true
			}

			if rule.Path != nil && *rule.Path != "" {
				model.Path = types.StringValue(*rule.Path)
				ruleHasData = true
			}

			if ruleHasData {
				rules = append(rules, model)
			}
		}
		if len(rules) > 0 {
			routes.Rules = rules
			hasData = true
		}
	}

	if !hasData {
		return nil
	}

	return routes
}

func flattenBulkSubscribe(api *catalyst_client.DaprSubscriptionSpecBulkSubscribe) *bulkSubscribeModel {
	if api == nil {
		return nil
	}

	model := &bulkSubscribeModel{
		Enabled:            types.BoolNull(),
		MaxMessagesCount:   types.Int64Null(),
		MaxAwaitDurationMs: types.Int64Null(),
	}
	hasData := false

	if api.Enabled != nil {
		model.Enabled = types.BoolValue(*api.Enabled)
		hasData = true
	}

	if api.MaxMessagesCount != nil {
		model.MaxMessagesCount = types.Int64Value(int64(*api.MaxMessagesCount))
		hasData = true
	}

	if api.MaxAwaitDurationMs != nil {
		model.MaxAwaitDurationMs = types.Int64Value(int64(*api.MaxAwaitDurationMs))
		hasData = true
	}

	if !hasData {
		return nil
	}

	return model
}

func mapToStringMap(ctx context.Context, m types.Map) (map[string]string, error) {
	if m.IsNull() || m.IsUnknown() {
		return map[string]string{}, nil
	}

	values := make(map[string]string)
	if diags := m.ElementsAs(ctx, &values, false); diags.HasError() {
		return nil, fmt.Errorf("failed to parse map: %v", diags.Errors())
	}
	return values, nil
}

func specModelIsEmpty(spec *specModel) bool {
	if spec == nil {
		return true
	}

	if spec.Routes != nil {
		return false
	}

	if spec.BulkSubscribe != nil {
		return false
	}

	if !spec.DeadLetterTopic.IsNull() && !spec.DeadLetterTopic.IsUnknown() {
		return false
	}

	if !spec.Metadata.IsNull() && !spec.Metadata.IsUnknown() {
		return false
	}

	return true
}

func applyExpandedSpec(target *catalyst_client.DaprSubscriptionSpec, extras *catalyst_client.DaprSubscriptionSpec) {
	if target == nil || extras == nil {
		return
	}

	if extras.Routes != nil {
		target.Routes = extras.Routes
	}

	if extras.BulkSubscribe != nil {
		target.BulkSubscribe = extras.BulkSubscribe
	}

	if extras.DeadLetterTopic != nil {
		target.DeadLetterTopic = extras.DeadLetterTopic
	}

	if extras.Metadata != nil {
		target.Metadata = extras.Metadata
	}
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

	subscription, err := client.GetSubscription(ctx, m.GetProjectName(), m.GetName(), &catalyst_client.DescribeDaprSubscriptionParams{})
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

		spec, err := flattenSubscriptionSpec(ctx, subscription.Spec)
		if err != nil {
			return err
		}
		if specModelIsEmpty(spec) {
			m.Spec = nil
		} else {
			m.Spec = spec
		}
	} else {
		m.Spec = nil
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
