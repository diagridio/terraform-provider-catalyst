package component

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// yamlSemanticEquality is a plan modifier that suppresses diffs when YAML values
// are semantically equal after normalization.
type yamlSemanticEquality struct{}

// YAMLSemanticEquality returns a plan modifier that uses semantic equality for YAML.
func YAMLSemanticEquality() planmodifier.String {
	return yamlSemanticEquality{}
}

func (m yamlSemanticEquality) Description(_ context.Context) string {
	return "Suppresses differences in YAML formatting that don't change semantic meaning."
}

func (m yamlSemanticEquality) MarkdownDescription(_ context.Context) string {
	return "Suppresses differences in YAML formatting that don't change semantic meaning."
}

func (m yamlSemanticEquality) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Only suppress diffs when both state and config values exist and are known
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() ||
		req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// If values are already equal, nothing to do
	if req.ConfigValue.Equal(req.StateValue) {
		return
	}

	// Check if values are semantically equal after normalization
	normalizedConfig, err := normalizeYAML(req.ConfigValue.ValueString())
	if err != nil {
		return
	}

	normalizedState, err := normalizeYAML(req.StateValue.ValueString())
	if err != nil {
		return
	}

	// If normalized values are equal, keep the state value to suppress the diff
	if normalizedConfig == normalizedState {
		resp.PlanValue = req.StateValue
	}
}
