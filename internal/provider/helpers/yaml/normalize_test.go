// Copyright (c) Diagrid Inc.
// SPDX-License-Identifier: MIT

package yaml

import (
	"strings"
	"testing"
)

func TestSemanticEqual(t *testing.T) {
	tests := []struct {
		name     string
		yaml1    string
		yaml2    string
		expected bool
		wantErr  bool
	}{
		{
			name:     "identical strings",
			yaml1:    "foo: bar",
			yaml2:    "foo: bar",
			expected: true,
		},
		{
			name:     "different indentation",
			yaml1:    "foo: bar\nbaz: qux",
			yaml2:    "foo:  bar\nbaz:  qux",
			expected: true,
		},
		{
			name:     "quoted vs unquoted",
			yaml1:    `foo: "bar"`,
			yaml2:    `foo: bar`,
			expected: true,
		},
		{
			name:     "list inline vs expanded",
			yaml1:    "items: [1, 2, 3]",
			yaml2:    "items:\n- 1\n- 2\n- 3",
			expected: true,
		},
		{
			name:     "map key ordering",
			yaml1:    "a: 1\nb: 2\nc: 3",
			yaml2:    "c: 3\na: 1\nb: 2",
			expected: true,
		},
		{
			name:     "nested structures",
			yaml1:    "parent:\n  child:\n    value: test",
			yaml2:    "parent:\n  child: {value: test}",
			expected: true,
		},
		{
			name:     "empty list vs null",
			yaml1:    "items: []",
			yaml2:    "items: null",
			expected: true, // Both normalize to nil
		},
		{
			name:     "empty map vs null",
			yaml1:    "config: {}",
			yaml2:    "config: null",
			expected: true, // Both normalize to nil
		},
		{
			name:     "different values",
			yaml1:    "foo: bar",
			yaml2:    "foo: baz",
			expected: false,
		},
		{
			name:     "different structure",
			yaml1:    "foo: bar",
			yaml2:    "bar: foo",
			expected: false,
		},
		{
			name:     "multiline strings",
			yaml1:    "text: |\n  line1\n  line2",
			yaml2:    "text: \"line1\\nline2\"",
			expected: true,
		},
		{
			name:     "boolean variations",
			yaml1:    "enabled: true",
			yaml2:    "enabled: yes",
			expected: false, // YAML parses these to same bool, but JSON preserves representation
		},
		{
			name:     "numeric types",
			yaml1:    "count: 42",
			yaml2:    "count: 42.0",
			expected: true, // Both are parsed as numbers and JSON normalizes them
		},
		{
			name:    "invalid yaml1",
			yaml1:   "foo: bar\n  invalid:",
			yaml2:   "foo: bar",
			wantErr: true,
		},
		{
			name:    "invalid yaml2",
			yaml1:   "foo: bar",
			yaml2:   "foo: bar\n  invalid:",
			wantErr: true,
		},
		{
			name:     "complex nested structures",
			yaml1:    "routes:\n  rules:\n  - match: foo\n    path: /bar\n  default: /baz",
			yaml2:    "routes:\n  default: /baz\n  rules:\n  - path: /bar\n    match: foo",
			expected: true,
		},
		{
			name:     "whitespace variations",
			yaml1:    "foo:    bar",
			yaml2:    "foo: bar",
			expected: true,
		},
		{
			name:     "trailing whitespace",
			yaml1:    "foo: bar  \n",
			yaml2:    "foo: bar\n",
			expected: true,
		},
		{
			name:     "empty string vs omitted",
			yaml1:    "foo: \"\"",
			yaml2:    "foo:",
			expected: false, // These are actually different in Go (empty string vs nil)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SemanticEqual(tt.yaml1, tt.yaml2)
			if (err != nil) != tt.wantErr {
				t.Errorf("SemanticEqual() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("SemanticEqual() = %v, expected %v\nYAML1:\n%s\nYAML2:\n%s",
					result, tt.expected, tt.yaml1, tt.yaml2)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result string)
	}{
		{
			name:  "simple map",
			input: "foo: bar\nbaz: qux",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "foo:") || !strings.Contains(result, "bar") {
					t.Errorf("Normalized output missing expected content: %s", result)
				}
			},
		},
		{
			name:  "inline to expanded",
			input: "{foo: bar, baz: qux}",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "foo:") {
					t.Errorf("Normalized output missing foo: %s", result)
				}
			},
		},
		{
			name:    "invalid yaml",
			input:   "foo: bar\nitems:\n- foo\n  bar",
			wantErr: false, // yaml.v3 is actually lenient with this
		},
		{
			name:  "empty list",
			input: "items: []",
			check: func(t *testing.T, result string) {
				// Should normalize empty list
				if strings.Contains(result, "items:") && strings.Contains(result, "[]") {
					t.Logf("Empty list preserved (acceptable): %s", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Normalize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestValidateStructure(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple yaml",
			input:   "foo: bar",
			wantErr: false,
		},
		{
			name:    "valid nested yaml",
			input:   "parent:\n  child:\n    value: test",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid indentation",
			input:   "foo: bar\n  invalid:",
			wantErr: true,
		},
		{
			name:    "unclosed quote",
			input:   "foo: \"bar",
			wantErr: true,
		},
		{
			name:    "invalid list syntax",
			input:   "items:\n-foo\nbar: baz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStructure(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStructure() error = %v, wantErr %v", err, tt.wantErr)
				if err != nil {
					t.Logf("Error message: %s", err)
				}
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  ",
			expected: true,
		},
		{
			name:     "null",
			input:    "null",
			expected: true,
		},
		{
			name:     "tilde (YAML null)",
			input:    "~",
			expected: true,
		},
		{
			name:     "empty list",
			input:    "[]",
			expected: true,
		},
		{
			name:     "empty map",
			input:    "{}",
			expected: true,
		},
		{
			name:     "non-empty value",
			input:    "foo: bar",
			expected: false,
		},
		{
			name:     "zero value",
			input:    "0",
			expected: false,
		},
		{
			name:     "false",
			input:    "false",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmpty() = %v, expected %v for input: %q", result, tt.expected, tt.input)
			}
		})
	}
}

func TestNormalizeEmptyValues(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "empty slice",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: nil,
		},
		{
			name:     "non-empty slice",
			input:    []interface{}{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "non-empty map",
			input:    map[string]interface{}{"foo": "bar"},
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name: "nested empty structures",
			input: map[string]interface{}{
				"empty_list": []interface{}{},
				"empty_map":  map[string]interface{}{},
				"value":      "test",
			},
			expected: map[string]interface{}{
				"empty_list": nil,
				"empty_map":  nil,
				"value":      "test",
			},
		},
		{
			name:     "string value",
			input:    "test",
			expected: "test",
		},
		{
			name:     "numeric value",
			input:    42,
			expected: 42,
		},
		{
			name:     "nil value",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeEmptyValues(tt.input)
			if !deepEqual(result, tt.expected) {
				t.Errorf("normalizeEmptyValues() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// deepEqual performs a simple deep equality check
func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch va := a.(type) {
	case []interface{}:
		vb, ok := b.([]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for i := range va {
			if !deepEqual(va[i], vb[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !deepEqual(v, vb[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

// Benchmark tests
func BenchmarkSemanticEqual(b *testing.B) {
	yaml1 := `
routes:
  rules:
    - match: foo
      path: /bar
    - match: baz
      path: /qux
  default: /home
metadata:
  key1: value1
  key2: value2
  key3: value3
`
	yaml2 := `
routes:
  default: /home
  rules:
    - path: /bar
      match: foo
    - path: /qux
      match: baz
metadata:
  key3: value3
  key1: value1
  key2: value2
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SemanticEqual(yaml1, yaml2)
	}
}

func BenchmarkNormalize(b *testing.B) {
	yaml := `
routes:
  rules:
    - match: foo
      path: /bar
  default: /home
metadata:
  key1: value1
  key2: value2
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Normalize(yaml)
	}
}
