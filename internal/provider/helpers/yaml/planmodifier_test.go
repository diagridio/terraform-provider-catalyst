// Copyright (c) Diagrid Inc.
// SPDX-License-Identifier: MIT

package yaml

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSemanticEquivalenceModifier(t *testing.T) {
	tests := []struct {
		name          string
		planValue     types.String
		stateValue    types.String
		expectedValue types.String // The value we expect in the response
		description   string
	}{
		{
			name:          "identical values - no change",
			planValue:     types.StringValue("foo: bar"),
			stateValue:    types.StringValue("foo: bar"),
			expectedValue: types.StringValue("foo: bar"),
			description:   "When values are identical, plan value should remain unchanged",
		},
		{
			name:          "semantically equal - different formatting",
			planValue:     types.StringValue("foo: bar\nbaz: qux"),
			stateValue:    types.StringValue("foo:  bar\nbaz:  qux"),
			expectedValue: types.StringValue("foo:  bar\nbaz:  qux"),
			description:   "When semantically equal, should use state value to suppress diff",
		},
		{
			name:          "different map ordering",
			planValue:     types.StringValue("a: 1\nb: 2"),
			stateValue:    types.StringValue("b: 2\na: 1"),
			expectedValue: types.StringValue("b: 2\na: 1"),
			description:   "Map key ordering shouldn't cause diff",
		},
		{
			name:          "quoted vs unquoted",
			planValue:     types.StringValue(`foo: "bar"`),
			stateValue:    types.StringValue(`foo: bar`),
			expectedValue: types.StringValue(`foo: bar`),
			description:   "Quote style differences shouldn't cause diff",
		},
		{
			name:          "different values - allow diff",
			planValue:     types.StringValue("foo: bar"),
			stateValue:    types.StringValue("foo: baz"),
			expectedValue: types.StringValue("foo: bar"),
			description:   "When values differ, plan value should be used",
		},
		{
			name:          "null plan value",
			planValue:     types.StringNull(),
			stateValue:    types.StringValue("foo: bar"),
			expectedValue: types.StringNull(),
			description:   "Null plan value should remain null",
		},
		{
			name:          "unknown plan value",
			planValue:     types.StringUnknown(),
			stateValue:    types.StringValue("foo: bar"),
			expectedValue: types.StringUnknown(),
			description:   "Unknown plan value should remain unknown",
		},
		{
			name:          "null state value",
			planValue:     types.StringValue("foo: bar"),
			stateValue:    types.StringNull(),
			expectedValue: types.StringValue("foo: bar"),
			description:   "Null state value should not affect plan value",
		},
		{
			name:          "invalid yaml in plan - allow through",
			planValue:     types.StringValue("foo: bar\n  invalid:"),
			stateValue:    types.StringValue("foo: bar"),
			expectedValue: types.StringValue("foo: bar\n  invalid:"),
			description:   "Invalid YAML should be allowed through for validation",
		},
		{
			name:          "complex nested structure",
			planValue:     types.StringValue("routes:\n  rules:\n  - match: foo\n    path: /bar\n  default: /baz"),
			stateValue:    types.StringValue("routes:\n  default: /baz\n  rules:\n  - path: /bar\n    match: foo"),
			expectedValue: types.StringValue("routes:\n  default: /baz\n  rules:\n  - path: /bar\n    match: foo"),
			description:   "Complex nested structures with different ordering should be equal",
		},
		{
			name:          "empty list vs null",
			planValue:     types.StringValue("items: []"),
			stateValue:    types.StringValue("items: null"),
			expectedValue: types.StringValue("items: null"),
			description:   "Empty list and null should be treated as equal",
		},
		{
			name:          "inline vs expanded list",
			planValue:     types.StringValue("items: [1, 2, 3]"),
			stateValue:    types.StringValue("items:\n- 1\n- 2\n- 3"),
			expectedValue: types.StringValue("items:\n- 1\n- 2\n- 3"),
			description:   "List syntax variations should be equal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifier := SemanticEquivalenceModifier()

			req := planmodifier.StringRequest{
				PlanValue:  tt.planValue,
				StateValue: tt.stateValue,
			}

			resp := &planmodifier.StringResponse{
				PlanValue: tt.planValue,
			}

			modifier.PlanModifyString(context.Background(), req, resp)

			if !resp.PlanValue.Equal(tt.expectedValue) {
				t.Errorf("PlanModifyString() result = %v, expected %v\n%s",
					resp.PlanValue, tt.expectedValue, tt.description)
			}
		})
	}
}

func TestSemanticEquivalenceModifier_Description(t *testing.T) {
	modifier := SemanticEquivalenceModifier()
	ctx := context.Background()

	desc := modifier.Description(ctx)
	if desc == "" {
		t.Error("Description() returned empty string")
	}

	mdDesc := modifier.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("MarkdownDescription() returned empty string")
	}

	// Both should be the same for this modifier
	if desc != mdDesc {
		t.Errorf("Description() = %q, but MarkdownDescription() = %q", desc, mdDesc)
	}
}

// TestSemanticEquivalenceModifier_Integration simulates a more realistic scenario
func TestSemanticEquivalenceModifier_Integration(t *testing.T) {
	// This test simulates what happens when a user reformats their YAML
	// between terraform apply runs
	originalYAML := `routes:
  rules:
    - match: "foo"
      path: "/bar"
  default: "/home"`

	reformattedYAML := `routes:
  default: /home
  rules:
  - match: foo
    path: /bar`

	modifier := SemanticEquivalenceModifier()

	req := planmodifier.StringRequest{
		PlanValue:  types.StringValue(reformattedYAML),
		StateValue: types.StringValue(originalYAML),
	}

	resp := &planmodifier.StringResponse{
		PlanValue: types.StringValue(reformattedYAML),
	}

	modifier.PlanModifyString(context.Background(), req, resp)

	// Should use state value since they're semantically equal
	if !resp.PlanValue.Equal(types.StringValue(originalYAML)) {
		t.Error("Expected plan modifier to suppress diff for semantically equal YAML")
		t.Logf("Plan value: %s", reformattedYAML)
		t.Logf("State value: %s", originalYAML)
		t.Logf("Result: %s", resp.PlanValue.ValueString())
	}
}

// Benchmark the plan modifier to ensure it's performant
func BenchmarkSemanticEquivalenceModifier(b *testing.B) {
	yaml1 := `routes:
  rules:
    - match: foo
      path: /bar
    - match: baz
      path: /qux
  default: /home
metadata:
  key1: value1
  key2: value2
  key3: value3`

	yaml2 := `routes:
  default: /home
  rules:
    - path: /bar
      match: foo
    - path: /qux
      match: baz
metadata:
  key3: value3
  key1: value1
  key2: value2`

	modifier := SemanticEquivalenceModifier()

	req := planmodifier.StringRequest{
		PlanValue:  types.StringValue(yaml1),
		StateValue: types.StringValue(yaml2),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := &planmodifier.StringResponse{
			PlanValue: req.PlanValue,
		}
		modifier.PlanModifyString(context.Background(), req, resp)
	}
}
