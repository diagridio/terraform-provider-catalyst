package resiliency

import (
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/helpers"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func resiliencySpecAttribute(required, computed bool) schema.SingleNestedAttribute {
	return helpers.SingleNestedAttr(
		"Dapr resiliency spec",
		required,
		computed,
		map[string]schema.Attribute{
			"policies": resiliencyPoliciesAttribute(computed),
			"targets":  resiliencyTargetsAttribute(computed),
		},
	)
}

func resiliencyPoliciesAttribute(computed bool) schema.SingleNestedAttribute {
	return helpers.SingleNestedAttr(
		"Resiliency policy definitions",
		false,
		computed,
		map[string]schema.Attribute{
			"timeouts":         helpers.MapStringAttr("Timeout policies keyed by name", false, computed),
			"retries":          resiliencyRetryPoliciesAttribute(computed),
			"circuit_breakers": resiliencyCircuitBreakerPoliciesAttribute(computed),
		},
	)
}

func resiliencyRetryPoliciesAttribute(computed bool) schema.MapNestedAttribute {
	return helpers.MapNestedAttr(
		"Retry policies keyed by name",
		false,
		computed,
		map[string]schema.Attribute{
			"duration":     helpers.StringAttr("Delay between retries (e.g. 5s)", computed),
			"max_interval": helpers.StringAttr("Maximum backoff interval", computed),
			"max_retries":  helpers.Int64Attr("Maximum retry attempts (-1 for infinite)", false, computed),
			"policy":       helpers.StringAttr("Retry policy type (constant or exponential)", computed),
		},
	)
}

func resiliencyCircuitBreakerPoliciesAttribute(computed bool) schema.MapNestedAttribute {
	return helpers.MapNestedAttr(
		"Circuit breaker policies keyed by name",
		false,
		computed,
		map[string]schema.Attribute{
			"interval":     helpers.StringAttr("Time window used to calculate statistics", computed),
			"max_requests": helpers.Int64Attr("Maximum requests allowed in half-open state", false, computed),
			"timeout":      helpers.StringAttr("Duration the circuit remains open", computed),
			"trip":         helpers.StringAttr("Condition that opens the circuit", computed),
		},
	)
}

func resiliencyTargetsAttribute(computed bool) schema.SingleNestedAttribute {
	return helpers.SingleNestedAttr(
		"Target assignments for policies",
		false,
		computed,
		map[string]schema.Attribute{
			"apps":       resiliencyTargetAppsAttribute(computed),
			"actors":     resiliencyTargetActorsAttribute(computed),
			"components": resiliencyTargetComponentsAttribute(computed),
		},
	)
}

func resiliencyTargetAppsAttribute(computed bool) schema.MapNestedAttribute {
	return helpers.MapNestedAttr(
		"Application policy bindings keyed by app ID",
		false,
		computed,
		map[string]schema.Attribute{
			"circuit_breaker":            helpers.StringAttr("Circuit breaker policy name", computed),
			"circuit_breaker_cache_size": helpers.Int64Attr("Size of the circuit breaker cache", false, computed),
			"retry":                      helpers.StringAttr("Retry policy name", computed),
			"timeout":                    helpers.StringAttr("Timeout policy name", computed),
		},
	)
}

func resiliencyTargetActorsAttribute(computed bool) schema.MapNestedAttribute {
	return helpers.MapNestedAttr(
		"Actor policy bindings keyed by actor type",
		false,
		computed,
		map[string]schema.Attribute{
			"circuit_breaker":            helpers.StringAttr("Circuit breaker policy name", computed),
			"circuit_breaker_cache_size": helpers.Int64Attr("Size of the circuit breaker cache", false, computed),
			"circuit_breaker_scope":      helpers.StringAttr("Scope used for the circuit breaker", computed),
			"retry":                      helpers.StringAttr("Retry policy name", computed),
			"timeout":                    helpers.StringAttr("Timeout policy name", computed),
		},
	)
}

func resiliencyTargetComponentsAttribute(computed bool) schema.MapNestedAttribute {
	return helpers.MapNestedAttr(
		"Component policy bindings keyed by component name",
		false,
		computed,
		map[string]schema.Attribute{
			"inbound":  resiliencyComponentDirectionAttribute("Policies applied to inbound operations", computed),
			"outbound": resiliencyComponentDirectionAttribute("Policies applied to outbound operations", computed),
		},
	)
}

func resiliencyComponentDirectionAttribute(desc string, computed bool) schema.SingleNestedAttribute {
	return helpers.SingleNestedAttr(
		desc,
		false,
		computed,
		map[string]schema.Attribute{
			"circuit_breaker": helpers.StringAttr("Circuit breaker policy name", computed),
			"retry":           helpers.StringAttr("Retry policy name", computed),
			"timeout":         helpers.StringAttr("Timeout policy name", computed),
		},
	)
}
