package configuration

import (
	"context"
	"fmt"

	catalyst_client "github.com/diagridio/cloudgrid/sdk/go/pkg/catalyst/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/diagridio/terraform-provider-catalyst/internal/catalyst"
)

func expandConfigurationSpec(ctx context.Context, spec *specModel) (*catalyst_client.DaprConfigurationSpec, error) {
	if spec == nil {
		return &catalyst_client.DaprConfigurationSpec{}, nil
	}

	apiSpec := &catalyst_client.DaprConfigurationSpec{}

	if spec.AccessControl != nil {
		ac, err := expandAccessControl(ctx, spec.AccessControl)
		if err != nil {
			return nil, err
		}
		apiSpec.AccessControl = ac
	}

	if spec.AppHTTPPipeline != nil {
		handlers := expandHandlers(spec.AppHTTPPipeline)
		if handlers != nil {
			apiSpec.AppHttpPipeline = &catalyst_client.ConfigurationSpecAppHttpPipeline{Handlers: handlers}
		}
	}

	if spec.HttpPipeline != nil {
		handlers := expandHandlers(spec.HttpPipeline)
		if handlers != nil {
			apiSpec.HttpPipeline = &catalyst_client.ConfigurationSpecHttpPipeline{Handlers: handlers}
		}
	}

	return apiSpec, nil
}

func expandAccessControl(ctx context.Context, model *accessControlModel) (*catalyst_client.ConfigurationSpecAccessControl, error) {
	if model == nil {
		return nil, nil
	}

	ac := &catalyst_client.ConfigurationSpecAccessControl{}

	if !model.DefaultAction.IsNull() && !model.DefaultAction.IsUnknown() {
		action := catalyst_client.ConfigurationSpecAccessControlDefaultAction(model.DefaultAction.ValueString())
		if action != "" {
			ac.DefaultAction = &action
		}
	}

	if !model.TrustDomain.IsNull() && !model.TrustDomain.IsUnknown() {
		trustDomain := model.TrustDomain.ValueString()
		if trustDomain != "" {
			ac.TrustDomain = &trustDomain
		}
	}

	if len(model.Policies) > 0 {
		policies := make([]catalyst_client.ConfigurationSpecAccessControlPolicy, 0, len(model.Policies))
		for _, policy := range model.Policies {
			p := catalyst_client.ConfigurationSpecAccessControlPolicy{}

			if !policy.AppID.IsNull() && !policy.AppID.IsUnknown() {
				appID := policy.AppID.ValueString()
				if appID != "" {
					p.AppId = &appID
				}
			}

			if !policy.DefaultAction.IsNull() && !policy.DefaultAction.IsUnknown() {
				action := catalyst_client.ConfigurationSpecAccessControlPolicyDefaultAction(policy.DefaultAction.ValueString())
				if string(action) != "" {
					p.DefaultAction = &action
				}
			}

			// Note: Namespace is not sent to the API as it's computed by the API itself

			if !policy.TrustDomain.IsNull() && !policy.TrustDomain.IsUnknown() {
				trustDomain := policy.TrustDomain.ValueString()
				if trustDomain != "" {
					p.TrustDomain = &trustDomain
				}
			}

			if len(policy.Operations) > 0 {
				operations := make([]struct {
					Action   *catalyst_client.ConfigurationSpecAccessControlPolicyOperationsAction `json:"action,omitempty"`
					HttpVerb *[]string                                                             `json:"httpVerb,omitempty"`
					Name     *string                                                               `json:"name,omitempty"`
				}, 0, len(policy.Operations))

				for _, op := range policy.Operations {
					operation := struct {
						Action   *catalyst_client.ConfigurationSpecAccessControlPolicyOperationsAction `json:"action,omitempty"`
						HttpVerb *[]string                                                             `json:"httpVerb,omitempty"`
						Name     *string                                                               `json:"name,omitempty"`
					}{}

					if !op.Name.IsNull() && !op.Name.IsUnknown() {
						name := op.Name.ValueString()
						if name != "" {
							operation.Name = &name
						}
					}

					if !op.Action.IsNull() && !op.Action.IsUnknown() {
						action := catalyst_client.ConfigurationSpecAccessControlPolicyOperationsAction(op.Action.ValueString())
						if string(action) != "" {
							operation.Action = &action
						}
					}

					httpVerbs, err := listToStringSlice(ctx, op.HTTPVerbs)
					if err != nil {
						return nil, err
					}
					if len(httpVerbs) > 0 {
						operation.HttpVerb = &httpVerbs
					}

					operations = append(operations, operation)
				}

				p.Operations = &operations
			}

			policies = append(policies, p)
		}

		ac.Policies = &policies
	}

	if ac.DefaultAction == nil && ac.Policies == nil && ac.TrustDomain == nil {
		return nil, nil
	}

	return ac, nil
}

func expandHandlers(model *pipelineModel) *[]catalyst_client.ConfigurationSpecHandler {
	if model == nil {
		return nil
	}

	if len(model.Handlers) == 0 {
		empty := []catalyst_client.ConfigurationSpecHandler{}
		return &empty
	}

	handlers := make([]catalyst_client.ConfigurationSpecHandler, 0, len(model.Handlers))
	for _, handler := range model.Handlers {
		h := catalyst_client.ConfigurationSpecHandler{}

		if !handler.Name.IsNull() && !handler.Name.IsUnknown() {
			name := handler.Name.ValueString()
			if name != "" {
				h.Name = &name
			}
		}

		if !handler.Type.IsNull() && !handler.Type.IsUnknown() {
			typ := handler.Type.ValueString()
			if typ != "" {
				h.Type = &typ
			}
		}

		handlers = append(handlers, h)
	}

	return &handlers
}

func flattenConfigurationSpec(ctx context.Context, apiSpec *catalyst_client.DaprConfigurationSpec) (*specModel, error) {
	spec := &specModel{}
	if apiSpec == nil {
		return spec, nil
	}

	if apiSpec.AccessControl != nil {
		ac, err := flattenAccessControl(ctx, apiSpec.AccessControl)
		if err != nil {
			return nil, err
		}
		spec.AccessControl = ac
	}

	if apiSpec.AppHttpPipeline != nil {
		spec.AppHTTPPipeline = flattenAppPipeline(apiSpec.AppHttpPipeline)
	}

	if apiSpec.HttpPipeline != nil {
		spec.HttpPipeline = flattenHTTPPipeline(apiSpec.HttpPipeline)
	}

	return spec, nil
}

func flattenAccessControl(ctx context.Context, api *catalyst_client.ConfigurationSpecAccessControl) (*accessControlModel, error) {
	if api == nil {
		return nil, nil
	}

	model := &accessControlModel{
		DefaultAction: types.StringNull(),
		TrustDomain:   types.StringNull(),
	}

	if api.DefaultAction != nil && string(*api.DefaultAction) != "" {
		model.DefaultAction = types.StringValue(string(*api.DefaultAction))
	}

	if api.TrustDomain != nil && *api.TrustDomain != "" {
		model.TrustDomain = types.StringValue(*api.TrustDomain)
	}

	if api.Policies != nil {
		for _, policy := range *api.Policies {
			p := accessControlPolicy{
				AppID:         types.StringNull(),
				DefaultAction: types.StringNull(),
				Namespace:     types.StringNull(),
				TrustDomain:   types.StringNull(),
			}

			if policy.AppId != nil && *policy.AppId != "" {
				p.AppID = types.StringValue(*policy.AppId)
			}

			if policy.DefaultAction != nil && string(*policy.DefaultAction) != "" {
				p.DefaultAction = types.StringValue(string(*policy.DefaultAction))
			}

			if policy.Namespace != nil && *policy.Namespace != "" {
				p.Namespace = types.StringValue(*policy.Namespace)
			}

			if policy.TrustDomain != nil && *policy.TrustDomain != "" {
				p.TrustDomain = types.StringValue(*policy.TrustDomain)
			}

			if policy.Operations != nil {
				for _, operation := range *policy.Operations {
					op := accessControlOperation{
						Name:      types.StringNull(),
						Action:    types.StringNull(),
						HTTPVerbs: types.ListNull(types.StringType),
					}

					if operation.Name != nil && *operation.Name != "" {
						op.Name = types.StringValue(*operation.Name)
					}

					if operation.Action != nil && string(*operation.Action) != "" {
						op.Action = types.StringValue(string(*operation.Action))
					}

					listValue, err := newStringListValue(ctx, operation.HttpVerb)
					if err != nil {
						return nil, err
					}
					op.HTTPVerbs = listValue

					p.Operations = append(p.Operations, op)
				}
			}

			model.Policies = append(model.Policies, p)
		}
	}

	if model.DefaultAction.IsNull() && model.TrustDomain.IsNull() && len(model.Policies) == 0 {
		return nil, nil
	}

	return model, nil
}

func flattenAppPipeline(api *catalyst_client.ConfigurationSpecAppHttpPipeline) *pipelineModel {
	if api == nil {
		return nil
	}
	return flattenHandlers(api.Handlers)
}

func flattenHTTPPipeline(api *catalyst_client.ConfigurationSpecHttpPipeline) *pipelineModel {
	if api == nil {
		return nil
	}
	return flattenHandlers(api.Handlers)
}

func flattenHandlers(handlers *[]catalyst_client.ConfigurationSpecHandler) *pipelineModel {
	if handlers == nil {
		return nil
	}

	model := &pipelineModel{}
	for _, handler := range *handlers {
		h := handlerModel{
			Name: types.StringNull(),
			Type: types.StringNull(),
		}

		if handler.Name != nil && *handler.Name != "" {
			h.Name = types.StringValue(*handler.Name)
		}

		if handler.Type != nil && *handler.Type != "" {
			h.Type = types.StringValue(*handler.Type)
		}

		model.Handlers = append(model.Handlers, h)
	}

	if len(model.Handlers) == 0 {
		return nil
	}

	return model
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

	configuration, err := client.GetConfiguration(ctx, m.GetProjectID(), m.GetName(), &catalyst_client.DescribeDaprConfigurationParams{})
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

	if configuration.Spec != nil {
		spec, err := flattenConfigurationSpec(ctx, configuration.Spec)
		if err != nil {
			return fmt.Errorf("error flattening configuration spec: %w", err)
		}
		m.Spec = spec
	} else {
		m.Spec = &specModel{}
	}

	if configuration.Status != nil && configuration.Status.Status != nil {
		m.SetStatus(*configuration.Status.Status)
	} else {
		m.SetStatus("")
	}

	m.Log(ctx, "read configuration model")

	return nil
}

func listToStringSlice(ctx context.Context, list types.List) ([]string, error) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var values []string
	diags := list.ElementsAs(ctx, &values, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse string list: %v", diags.Errors())
	}
	return values, nil
}

func newStringListValue(ctx context.Context, values *[]string) (types.List, error) {
	if values == nil {
		return types.ListNull(types.StringType), nil
	}

	list, diags := types.ListValueFrom(ctx, types.StringType, *values)
	if diags.HasError() {
		return types.ListNull(types.StringType), fmt.Errorf("failed to convert string slice: %v", diags.Errors())
	}
	return list, nil
}

// mergeSpecWithPlanned merges the planned spec values with the read spec values.
// This is necessary because the API may not return all fields that were set in the plan,
// causing Terraform to detect inconsistencies between plan and state.
// If a field is nil in the read spec but was set in the planned spec, we use the planned value.
func mergeSpecWithPlanned(readSpec, plannedSpec *specModel) {
	if readSpec == nil || plannedSpec == nil {
		return
	}

	// Preserve planned AccessControl config if it wasn't returned by the API
	if readSpec.AccessControl == nil && plannedSpec.AccessControl != nil {
		readSpec.AccessControl = plannedSpec.AccessControl
	}

	// Preserve planned AppHTTPPipeline config if it wasn't returned by the API
	if readSpec.AppHTTPPipeline == nil && plannedSpec.AppHTTPPipeline != nil {
		readSpec.AppHTTPPipeline = plannedSpec.AppHTTPPipeline
	}

	// Preserve planned HttpPipeline config if it wasn't returned by the API
	if readSpec.HttpPipeline == nil && plannedSpec.HttpPipeline != nil {
		readSpec.HttpPipeline = plannedSpec.HttpPipeline
	}
}
