package subscription

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"
)

// yamlEquivalencePlanModifier checks if two YAML strings are semantically equivalent.
type yamlEquivalencePlanModifier struct{}

func YAMLEquivalence() planmodifier.String {
	return yamlEquivalencePlanModifier{}
}

func (m yamlEquivalencePlanModifier) Description(_ context.Context) string {
	return "Suppresses diff when YAML content is semantically equivalent"
}

func (m yamlEquivalencePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Suppresses diff when YAML content is semantically equivalent"
}

func (m yamlEquivalencePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If either value is null or unknown, use default behavior
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// Parse both YAML strings into generic structures
	var planData, stateData interface{}

	if err := yaml.Unmarshal([]byte(req.PlanValue.ValueString()), &planData); err != nil {
		// If parsing fails, use default behavior
		return
	}

	if err := yaml.Unmarshal([]byte(req.StateValue.ValueString()), &stateData); err != nil {
		// If parsing fails, use default behavior
		return
	}

	// Compare via JSON marshaling (which normalizes the structures)
	planJSON, err := json.Marshal(planData)
	if err != nil {
		return
	}

	stateJSON, err := json.Marshal(stateData)
	if err != nil {
		return
	}

	// If JSON representations are equal, the YAML is semantically equivalent
	if string(planJSON) == string(stateJSON) {
		// Use the state value to avoid unnecessary updates
		resp.PlanValue = types.StringValue(req.StateValue.ValueString())
	}
}
