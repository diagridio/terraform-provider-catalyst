// Copyright (c) Diagrid Inc.
// SPDX-License-Identifier: MIT

package yaml

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// SemanticEquivalenceModifier returns a plan modifier that compares YAML strings semantically.
//
// This plan modifier prevents spurious diffs when YAML content is semantically equivalent
// but formatted differently. It handles variations in:
//   - Indentation and whitespace
//   - Quote styles (quoted vs unquoted strings)
//   - Map key ordering
//   - List syntax (inline vs expanded)
//   - Empty values (empty list vs null, etc.)
//
// Inspired by the Kubernetes Terraform provider's approach to YAML handling.
//
// Example usage:
//
//	"spec": schema.StringAttribute{
//	    MarkdownDescription: "Configuration spec in YAML format",
//	    Required:            true,
//	    PlanModifiers: []planmodifier.String{
//	        yamlhelpers.SemanticEquivalenceModifier(),
//	    },
//	}
func SemanticEquivalenceModifier() planmodifier.String {
	return semanticEquivalenceModifier{}
}

type semanticEquivalenceModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m semanticEquivalenceModifier) Description(_ context.Context) string {
	return "Suppresses differences in YAML formatting that don't change semantic meaning"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m semanticEquivalenceModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

// PlanModifyString implements the plan modifier logic.
//
// This follows the pattern from the Kubernetes provider:
// 1. Check if either value is null or unknown (use default behavior)
// 2. If values are string-equal, no modification needed
// 3. Parse and compare semantically using canonical JSON representation
// 4. If semantically equal, use state value to suppress diff
func (m semanticEquivalenceModifier) PlanModifyString(
	_ context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	// Don't modify if either value is null or unknown
	// This allows Terraform to handle these cases with its default behavior
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() ||
		req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// If values are already string-equal, no modification needed
	// This is a fast path that avoids unnecessary parsing
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	// Compare semantically using our canonical normalization
	equal, err := SemanticEqual(
		req.PlanValue.ValueString(),
		req.StateValue.ValueString(),
	)
	if err != nil {
		// If comparison fails (e.g., invalid YAML), let default behavior handle it
		// This will typically result in Terraform showing the diff and letting
		// the resource's validation catch the invalid YAML
		return
	}

	if equal {
		// Values are semantically equal, use state value to suppress diff
		// This preserves the user's original formatting from the state
		resp.PlanValue = req.StateValue
	}
}
