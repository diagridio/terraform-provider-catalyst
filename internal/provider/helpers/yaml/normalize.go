// Copyright (c) Diagrid Inc.
// SPDX-License-Identifier: MIT

package yaml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// SemanticEqual compares two YAML strings for semantic equality.
// It normalizes both YAML strings and compares their canonical JSON representations.
// This ensures that formatting differences (indentation, quotes, key ordering, etc.)
// don't cause false positives in diff detection.
func SemanticEqual(yaml1, yaml2 string) (bool, error) {
	// Parse both YAML strings
	var data1, data2 interface{}

	if err := yaml.Unmarshal([]byte(yaml1), &data1); err != nil {
		return false, fmt.Errorf("failed to parse first YAML: %w", err)
	}

	if err := yaml.Unmarshal([]byte(yaml2), &data2); err != nil {
		return false, fmt.Errorf("failed to parse second YAML: %w", err)
	}

	// Normalize empty values to ensure consistent comparison
	data1 = normalizeEmptyValues(data1)
	data2 = normalizeEmptyValues(data2)

	// Convert to canonical JSON for comparison
	json1, err := marshalCanonicalJSON(data1)
	if err != nil {
		return false, fmt.Errorf("failed to marshal first YAML to JSON: %w", err)
	}

	json2, err := marshalCanonicalJSON(data2)
	if err != nil {
		return false, fmt.Errorf("failed to marshal second YAML to JSON: %w", err)
	}

	return bytes.Equal(json1, json2), nil
}

// Normalize converts a YAML string to a normalized form by unmarshaling and remarshaling.
// This is useful for storing a canonical representation in state.
func Normalize(yamlStr string) (string, error) {
	var data interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	data = normalizeEmptyValues(data)

	normalized, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(normalized), nil
}

// marshalCanonicalJSON creates a canonical JSON representation.
// This uses compact JSON format to eliminate whitespace variations.
func marshalCanonicalJSON(v interface{}) ([]byte, error) {
	// Marshal to JSON
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Compact to remove whitespace variations
	var buf bytes.Buffer
	if err := json.Compact(&buf, jsonBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// normalizeEmptyValues converts empty slices and maps to nil for consistent comparison.
// This is inspired by the Kubernetes provider's approach to handling empty vs. null values.
//
// Note: This treats empty collections the same as nil, which matches most API behavior.
// If your API distinguishes between empty and nil, this function may need adjustment.
func normalizeEmptyValues(data interface{}) interface{} {
	switch v := data.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil
		}
		// Recursively normalize nested structures
		for i := range v {
			v[i] = normalizeEmptyValues(v[i])
		}
		return v
	case map[string]interface{}:
		if len(v) == 0 {
			return nil
		}
		// Recursively normalize nested structures
		for k := range v {
			v[k] = normalizeEmptyValues(v[k])
		}
		return v
	default:
		return v
	}
}

// ValidateStructure performs basic structural validation on YAML.
// It checks for common issues and provides helpful error messages with context.
func ValidateStructure(yamlStr string) error {
	if yamlStr == "" {
		return fmt.Errorf("YAML string is empty")
	}

	var data interface{}
	err := yaml.Unmarshal([]byte(yamlStr), &data)
	if err != nil {
		// Try to provide context for the error
		if yerr, ok := err.(*yaml.TypeError); ok {
			return fmt.Errorf("YAML type error: %s\n%s", yerr.Errors, getYAMLContext(yamlStr, yerr.Errors))
		}
		// For syntax errors, try to extract line information
		return enhanceYAMLError(err, yamlStr)
	}

	return nil
}

// enhanceYAMLError attempts to add context to YAML parsing errors.
func enhanceYAMLError(err error, yamlStr string) error {
	errStr := err.Error()

	// Try to extract line number from error message
	// Common format: "yaml: line X: error message"
	if idx := strings.Index(errStr, "line "); idx != -1 {
		// Extract and provide context
		return fmt.Errorf("%s\n%s", errStr, extractLineContext(yamlStr, errStr))
	}

	return err
}

// extractLineContext tries to extract context lines from YAML based on error message.
func extractLineContext(yamlStr, errMsg string) string {
	// Simple approach: show first few lines if we can't parse line number
	lines := strings.Split(yamlStr, "\n")
	if len(lines) <= 5 {
		return fmt.Sprintf("YAML content:\n%s", yamlStr)
	}

	return fmt.Sprintf("First few lines:\n%s\n...", strings.Join(lines[:5], "\n"))
}

// getYAMLContext provides context for type errors.
func getYAMLContext(yamlStr string, errors []string) string {
	if len(errors) == 0 {
		return ""
	}

	var context strings.Builder
	context.WriteString("Context:\n")
	for i, err := range errors {
		if i >= 3 {
			context.WriteString("... and more\n")
			break
		}
		context.WriteString(fmt.Sprintf("  - %s\n", err))
	}

	return context.String()
}

// IsEmpty checks if a YAML string represents an empty or null value.
func IsEmpty(yamlStr string) bool {
	yamlStr = strings.TrimSpace(yamlStr)
	if yamlStr == "" || yamlStr == "null" || yamlStr == "~" {
		return true
	}

	// Try to parse and check if it's an empty structure
	var data interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &data); err != nil {
		return false
	}

	// Check if normalized value is nil
	normalized := normalizeEmptyValues(data)
	return normalized == nil
}
