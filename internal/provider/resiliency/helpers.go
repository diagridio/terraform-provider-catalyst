package resiliency

import (
	"context"
	"fmt"

	catalyst_client "github.com/diagridio/cloudgrid/sdk/go/pkg/catalyst/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

// toAPIScopes converts a Terraform List to API scopes.
func toAPIScopes(ctx context.Context, scopesList types.List) *[]string {
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

func expandResiliencySpec(ctx context.Context, model *specModel) (*catalyst_client.DaprResiliencySpec, error) {
	if model == nil {
		return &catalyst_client.DaprResiliencySpec{}, nil
	}

	api := &catalyst_client.DaprResiliencySpec{}

	if model.Policies != nil {
		policies, err := expandResiliencyPolicies(ctx, model.Policies)
		if err != nil {
			return nil, err
		}
		api.Policies = policies
	}

	if model.Targets != nil {
		targets := expandResiliencyTargets(model.Targets)
		api.Targets = targets
	}

	return api, nil
}

func expandResiliencyPolicies(ctx context.Context, model *policiesModel) (*catalyst_client.DaprResiliencySpecPolicies, error) {
	if model == nil {
		return nil, nil
	}

	policies := &catalyst_client.DaprResiliencySpecPolicies{}
	hasData := false

	if !model.Timeouts.IsNull() && !model.Timeouts.IsUnknown() {
		timeouts, err := mapToStringMap(ctx, model.Timeouts)
		if err != nil {
			return nil, fmt.Errorf("failed to expand timeouts: %w", err)
		}
		if len(timeouts) > 0 {
			policies.Timeouts = &timeouts
			hasData = true
		}
	}

	if len(model.Retries) > 0 {
		retries := make(map[string]catalyst_client.DaprResiliencySpecPoliciesRetries, len(model.Retries))
		for name, retryModel := range model.Retries {
			if retry, ok := expandRetryPolicy(retryModel); ok {
				retries[name] = retry
			}
		}
		if len(retries) > 0 {
			policies.Retries = &retries
			hasData = true
		}
	}

	if len(model.CircuitBreakers) > 0 {
		cbs := make(map[string]catalyst_client.DaprResiliencySpecPoliciesCircuitBreakers, len(model.CircuitBreakers))
		for name, cbModel := range model.CircuitBreakers {
			if cb, ok := expandCircuitBreaker(cbModel); ok {
				cbs[name] = cb
			}
		}
		if len(cbs) > 0 {
			policies.CircuitBreakers = &cbs
			hasData = true
		}
	}

	if !hasData {
		return nil, nil
	}

	return policies, nil
}

func expandRetryPolicy(model retryPolicyModel) (catalyst_client.DaprResiliencySpecPoliciesRetries, bool) {
	retry := catalyst_client.DaprResiliencySpecPoliciesRetries{}
	hasValue := false

	if !model.Duration.IsNull() && !model.Duration.IsUnknown() {
		val := model.Duration.ValueString()
		if val != "" {
			retry.Duration = &val
			hasValue = true
		}
	}

	if !model.MaxInterval.IsNull() && !model.MaxInterval.IsUnknown() {
		val := model.MaxInterval.ValueString()
		if val != "" {
			retry.MaxInterval = &val
			hasValue = true
		}
	}

	if !model.MaxRetries.IsNull() && !model.MaxRetries.IsUnknown() {
		val := int(model.MaxRetries.ValueInt64())
		retry.MaxRetries = &val
		hasValue = true
	}

	if !model.Policy.IsNull() && !model.Policy.IsUnknown() {
		val := model.Policy.ValueString()
		if val != "" {
			retry.Policy = &val
			hasValue = true
		}
	}

	return retry, hasValue
}

func expandCircuitBreaker(model circuitBreakerModel) (catalyst_client.DaprResiliencySpecPoliciesCircuitBreakers, bool) {
	cb := catalyst_client.DaprResiliencySpecPoliciesCircuitBreakers{}
	hasValue := false

	if !model.Interval.IsNull() && !model.Interval.IsUnknown() {
		val := model.Interval.ValueString()
		if val != "" {
			cb.Interval = &val
			hasValue = true
		}
	}

	if !model.MaxRequests.IsNull() && !model.MaxRequests.IsUnknown() {
		val := int(model.MaxRequests.ValueInt64())
		cb.MaxRequests = &val
		hasValue = true
	}

	if !model.Timeout.IsNull() && !model.Timeout.IsUnknown() {
		val := model.Timeout.ValueString()
		if val != "" {
			cb.Timeout = &val
			hasValue = true
		}
	}

	if !model.Trip.IsNull() && !model.Trip.IsUnknown() {
		val := model.Trip.ValueString()
		if val != "" {
			cb.Trip = &val
			hasValue = true
		}
	}

	return cb, hasValue
}

func expandResiliencyTargets(model *targetsModel) *catalyst_client.DaprResiliencySpecTargets {
	if model == nil {
		return nil
	}

	targets := &catalyst_client.DaprResiliencySpecTargets{}
	hasData := false

	if len(model.Apps) > 0 {
		apps := make(map[string]catalyst_client.DaprResiliencySpecTargetsApps, len(model.Apps))
		for name, appModel := range model.Apps {
			if endpoint, ok := expandEndpointPolicy(appModel); ok {
				apps[name] = endpoint
			}
		}
		if len(apps) > 0 {
			targets.Apps = &apps
			hasData = true
		}
	}

	if len(model.Actors) > 0 {
		actors := make(map[string]catalyst_client.DaprResiliencySpecTargetsActors, len(model.Actors))
		for name, actorModel := range model.Actors {
			if actor, ok := expandActorPolicy(actorModel); ok {
				actors[name] = actor
			}
		}
		if len(actors) > 0 {
			targets.Actors = &actors
			hasData = true
		}
	}

	if len(model.Components) > 0 {
		components := make(map[string]catalyst_client.DaprResiliencySpecTargetsComponents, len(model.Components))
		for name, componentModel := range model.Components {
			if component, ok := expandComponentPolicy(componentModel); ok {
				components[name] = component
			}
		}
		if len(components) > 0 {
			targets.Components = &components
			hasData = true
		}
	}

	if !hasData {
		return nil
	}

	return targets
}

func expandEndpointPolicy(model endpointPolicyModel) (catalyst_client.DaprResiliencySpecTargetsApps, bool) {
	policy := catalyst_client.DaprResiliencySpecTargetsApps{}
	hasValue := false

	if !model.CircuitBreaker.IsNull() && !model.CircuitBreaker.IsUnknown() {
		val := model.CircuitBreaker.ValueString()
		if val != "" {
			policy.CircuitBreaker = &val
			hasValue = true
		}
	}

	if !model.CircuitBreakerCacheSize.IsNull() && !model.CircuitBreakerCacheSize.IsUnknown() {
		val := int(model.CircuitBreakerCacheSize.ValueInt64())
		policy.CircuitBreakerCacheSize = &val
		hasValue = true
	}

	if !model.Retry.IsNull() && !model.Retry.IsUnknown() {
		val := model.Retry.ValueString()
		if val != "" {
			policy.Retry = &val
			hasValue = true
		}
	}

	if !model.Timeout.IsNull() && !model.Timeout.IsUnknown() {
		val := model.Timeout.ValueString()
		if val != "" {
			policy.Timeout = &val
			hasValue = true
		}
	}

	return policy, hasValue
}

func expandActorPolicy(model actorPolicyModel) (catalyst_client.DaprResiliencySpecTargetsActors, bool) {
	policy := catalyst_client.DaprResiliencySpecTargetsActors{}
	hasValue := false

	if !model.CircuitBreaker.IsNull() && !model.CircuitBreaker.IsUnknown() {
		val := model.CircuitBreaker.ValueString()
		if val != "" {
			policy.CircuitBreaker = &val
			hasValue = true
		}
	}

	if !model.CircuitBreakerCacheSize.IsNull() && !model.CircuitBreakerCacheSize.IsUnknown() {
		val := int(model.CircuitBreakerCacheSize.ValueInt64())
		policy.CircuitBreakerCacheSize = &val
		hasValue = true
	}

	if !model.CircuitBreakerScope.IsNull() && !model.CircuitBreakerScope.IsUnknown() {
		val := model.CircuitBreakerScope.ValueString()
		if val != "" {
			policy.CircuitBreakerScope = &val
			hasValue = true
		}
	}

	if !model.Retry.IsNull() && !model.Retry.IsUnknown() {
		val := model.Retry.ValueString()
		if val != "" {
			policy.Retry = &val
			hasValue = true
		}
	}

	if !model.Timeout.IsNull() && !model.Timeout.IsUnknown() {
		val := model.Timeout.ValueString()
		if val != "" {
			policy.Timeout = &val
			hasValue = true
		}
	}

	return policy, hasValue
}

func expandComponentPolicy(model componentPolicyModel) (catalyst_client.DaprResiliencySpecTargetsComponents, bool) {
	policy := catalyst_client.DaprResiliencySpecTargetsComponents{}
	hasValue := false

	if inbound := expandComponentDirectionPolicy(model.Inbound); inbound != nil {
		policy.Inbound = inbound
		hasValue = true
	}

	if outbound := expandComponentDirectionPolicy(model.Outbound); outbound != nil {
		policy.Outbound = outbound
		hasValue = true
	}

	return policy, hasValue
}

func expandComponentDirectionPolicy(model *componentDirectionPolicyModel) *catalyst_client.DaprResiliencySpecTargetsComponentsInbound {
	if model == nil {
		return nil
	}

	direction := &catalyst_client.DaprResiliencySpecTargetsComponentsInbound{}
	hasValue := false

	if !model.CircuitBreaker.IsNull() && !model.CircuitBreaker.IsUnknown() {
		val := model.CircuitBreaker.ValueString()
		if val != "" {
			direction.CircuitBreaker = &val
			hasValue = true
		}
	}

	if !model.Retry.IsNull() && !model.Retry.IsUnknown() {
		val := model.Retry.ValueString()
		if val != "" {
			direction.Retry = &val
			hasValue = true
		}
	}

	if !model.Timeout.IsNull() && !model.Timeout.IsUnknown() {
		val := model.Timeout.ValueString()
		if val != "" {
			direction.Timeout = &val
			hasValue = true
		}
	}

	if !hasValue {
		return nil
	}

	return direction
}

func flattenResiliencySpec(ctx context.Context, api *catalyst_client.DaprResiliencySpec) (*specModel, error) {
	model := &specModel{}
	if api == nil {
		return model, nil
	}

	if api.Policies != nil {
		policies, err := flattenResiliencyPolicies(ctx, api.Policies)
		if err != nil {
			return nil, err
		}
		model.Policies = policies
	}

	if api.Targets != nil {
		targets := flattenResiliencyTargets(api.Targets)
		model.Targets = targets
	}

	return model, nil
}

func flattenResiliencyPolicies(ctx context.Context, api *catalyst_client.DaprResiliencySpecPolicies) (*policiesModel, error) {
	if api == nil {
		return nil, nil
	}

	model := &policiesModel{
		Timeouts: types.MapNull(types.StringType),
	}
	hasData := false

	if api.Timeouts != nil && len(*api.Timeouts) > 0 {
		timeoutsValue, diags := types.MapValueFrom(ctx, types.StringType, *api.Timeouts)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to flatten timeouts: %v", diags.Errors())
		}
		model.Timeouts = timeoutsValue
		hasData = true
	}

	if api.Retries != nil && len(*api.Retries) > 0 {
		retries := make(map[string]retryPolicyModel, len(*api.Retries))
		for name, retry := range *api.Retries {
			if retryModel, ok := flattenRetryPolicy(retry); ok {
				retries[name] = retryModel
			}
		}
		if len(retries) > 0 {
			model.Retries = retries
			hasData = true
		}
	}

	if api.CircuitBreakers != nil && len(*api.CircuitBreakers) > 0 {
		cbs := make(map[string]circuitBreakerModel, len(*api.CircuitBreakers))
		for name, cb := range *api.CircuitBreakers {
			if cbModel, ok := flattenCircuitBreaker(cb); ok {
				cbs[name] = cbModel
			}
		}
		if len(cbs) > 0 {
			model.CircuitBreakers = cbs
			hasData = true
		}
	}

	if !hasData {
		return nil, nil
	}

	return model, nil
}

func flattenRetryPolicy(api catalyst_client.DaprResiliencySpecPoliciesRetries) (retryPolicyModel, bool) {
	model := retryPolicyModel{
		Duration:    types.StringNull(),
		MaxInterval: types.StringNull(),
		MaxRetries:  types.Int64Null(),
		Policy:      types.StringNull(),
	}
	hasValue := false

	if api.Duration != nil && *api.Duration != "" {
		model.Duration = types.StringValue(*api.Duration)
		hasValue = true
	}

	if api.MaxInterval != nil && *api.MaxInterval != "" {
		model.MaxInterval = types.StringValue(*api.MaxInterval)
		hasValue = true
	}

	if api.MaxRetries != nil {
		model.MaxRetries = types.Int64Value(int64(*api.MaxRetries))
		hasValue = true
	}

	if api.Policy != nil && *api.Policy != "" {
		model.Policy = types.StringValue(*api.Policy)
		hasValue = true
	}

	return model, hasValue
}

func flattenCircuitBreaker(api catalyst_client.DaprResiliencySpecPoliciesCircuitBreakers) (circuitBreakerModel, bool) {
	model := circuitBreakerModel{
		Interval:    types.StringNull(),
		MaxRequests: types.Int64Null(),
		Timeout:     types.StringNull(),
		Trip:        types.StringNull(),
	}
	hasValue := false

	if api.Interval != nil && *api.Interval != "" {
		model.Interval = types.StringValue(*api.Interval)
		hasValue = true
	}

	if api.MaxRequests != nil {
		model.MaxRequests = types.Int64Value(int64(*api.MaxRequests))
		hasValue = true
	}

	if api.Timeout != nil && *api.Timeout != "" {
		model.Timeout = types.StringValue(*api.Timeout)
		hasValue = true
	}

	if api.Trip != nil && *api.Trip != "" {
		model.Trip = types.StringValue(*api.Trip)
		hasValue = true
	}

	return model, hasValue
}

func flattenResiliencyTargets(api *catalyst_client.DaprResiliencySpecTargets) *targetsModel {
	if api == nil {
		return nil
	}

	model := &targetsModel{}
	hasData := false

	if api.Apps != nil && len(*api.Apps) > 0 {
		apps := make(map[string]endpointPolicyModel, len(*api.Apps))
		for name, endpoint := range *api.Apps {
			if endpointModel, ok := flattenEndpointPolicy(endpoint); ok {
				apps[name] = endpointModel
			}
		}
		if len(apps) > 0 {
			model.Apps = apps
			hasData = true
		}
	}

	if api.Actors != nil && len(*api.Actors) > 0 {
		actors := make(map[string]actorPolicyModel, len(*api.Actors))
		for name, actor := range *api.Actors {
			if actorModel, ok := flattenActorPolicy(actor); ok {
				actors[name] = actorModel
			}
		}
		if len(actors) > 0 {
			model.Actors = actors
			hasData = true
		}
	}

	if api.Components != nil && len(*api.Components) > 0 {
		components := make(map[string]componentPolicyModel, len(*api.Components))
		for name, component := range *api.Components {
			if componentModel, ok := flattenComponentPolicy(component); ok {
				components[name] = componentModel
			}
		}
		if len(components) > 0 {
			model.Components = components
			hasData = true
		}
	}

	if !hasData {
		return nil
	}

	return model
}

func flattenEndpointPolicy(api catalyst_client.DaprResiliencySpecTargetsApps) (endpointPolicyModel, bool) {
	model := endpointPolicyModel{
		CircuitBreaker:          types.StringNull(),
		CircuitBreakerCacheSize: types.Int64Null(),
		Retry:                   types.StringNull(),
		Timeout:                 types.StringNull(),
	}
	hasValue := false

	if api.CircuitBreaker != nil && *api.CircuitBreaker != "" {
		model.CircuitBreaker = types.StringValue(*api.CircuitBreaker)
		hasValue = true
	}

	if api.CircuitBreakerCacheSize != nil {
		model.CircuitBreakerCacheSize = types.Int64Value(int64(*api.CircuitBreakerCacheSize))
		hasValue = true
	}

	if api.Retry != nil && *api.Retry != "" {
		model.Retry = types.StringValue(*api.Retry)
		hasValue = true
	}

	if api.Timeout != nil && *api.Timeout != "" {
		model.Timeout = types.StringValue(*api.Timeout)
		hasValue = true
	}

	return model, hasValue
}

func flattenActorPolicy(api catalyst_client.DaprResiliencySpecTargetsActors) (actorPolicyModel, bool) {
	model := actorPolicyModel{
		CircuitBreaker:          types.StringNull(),
		CircuitBreakerCacheSize: types.Int64Null(),
		CircuitBreakerScope:     types.StringNull(),
		Retry:                   types.StringNull(),
		Timeout:                 types.StringNull(),
	}
	hasValue := false

	if api.CircuitBreaker != nil && *api.CircuitBreaker != "" {
		model.CircuitBreaker = types.StringValue(*api.CircuitBreaker)
		hasValue = true
	}

	if api.CircuitBreakerCacheSize != nil {
		model.CircuitBreakerCacheSize = types.Int64Value(int64(*api.CircuitBreakerCacheSize))
		hasValue = true
	}

	if api.CircuitBreakerScope != nil && *api.CircuitBreakerScope != "" {
		model.CircuitBreakerScope = types.StringValue(*api.CircuitBreakerScope)
		hasValue = true
	}

	if api.Retry != nil && *api.Retry != "" {
		model.Retry = types.StringValue(*api.Retry)
		hasValue = true
	}

	if api.Timeout != nil && *api.Timeout != "" {
		model.Timeout = types.StringValue(*api.Timeout)
		hasValue = true
	}

	return model, hasValue
}

func flattenComponentPolicy(api catalyst_client.DaprResiliencySpecTargetsComponents) (componentPolicyModel, bool) {
	model := componentPolicyModel{}
	hasValue := false

	if inbound := flattenComponentDirectionPolicy(api.Inbound); inbound != nil {
		model.Inbound = inbound
		hasValue = true
	}

	if outbound := flattenComponentDirectionPolicy(api.Outbound); outbound != nil {
		model.Outbound = outbound
		hasValue = true
	}

	return model, hasValue
}

func flattenComponentDirectionPolicy(api *catalyst_client.DaprResiliencySpecTargetsComponentsInbound) *componentDirectionPolicyModel {
	if api == nil {
		return nil
	}

	model := &componentDirectionPolicyModel{
		CircuitBreaker: types.StringNull(),
		Retry:          types.StringNull(),
		Timeout:        types.StringNull(),
	}
	hasValue := false

	if api.CircuitBreaker != nil && *api.CircuitBreaker != "" {
		model.CircuitBreaker = types.StringValue(*api.CircuitBreaker)
		hasValue = true
	}

	if api.Retry != nil && *api.Retry != "" {
		model.Retry = types.StringValue(*api.Retry)
		hasValue = true
	}

	if api.Timeout != nil && *api.Timeout != "" {
		model.Timeout = types.StringValue(*api.Timeout)
		hasValue = true
	}

	if !hasValue {
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

func read(ctx context.Context,
	client catalyst.Client,
	m *model,
) error {
	tflog.Debug(ctx, "reading resiliency",
		map[string]interface{}{
			"project_id": m.GetProjectID(),
			"name":       m.GetName(),
		})

	resiliency, err := client.GetResiliency(ctx, m.GetProjectID(), m.GetName(), &catalyst_client.DescribeDaprResiliencyParams{})
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

	if resiliency.Status != nil {
		m.SetStatus(resiliency.GetStatus())
	} else {
		m.SetStatus("")
	}

	spec, err := flattenResiliencySpec(ctx, resiliency.Spec)
	if err != nil {
		return fmt.Errorf("error flattening resiliency spec: %w", err)
	}
	m.Spec = spec

	m.Log(ctx, "read resiliency model")

	return nil
}
