package resiliency

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringAttr(desc string, required, computed bool) schema.StringAttribute {
	req, opt, comp := attributeModes(required, computed)
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

func boolAttr(desc string, required, computed bool) schema.BoolAttribute {
	req, opt, comp := attributeModes(required, computed)
	return schema.BoolAttribute{
		MarkdownDescription: desc,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

func int64Attr(desc string, required, computed bool) schema.Int64Attribute {
	req, opt, comp := attributeModes(required, computed)
	return schema.Int64Attribute{
		MarkdownDescription: desc,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

func mapStringAttr(desc string, required, computed bool) schema.MapAttribute {
	req, opt, comp := attributeModes(required, computed)
	return schema.MapAttribute{
		MarkdownDescription: desc,
		ElementType:         types.StringType,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

func listNestedAttr(desc string, required, computed bool, attrs map[string]schema.Attribute) schema.ListNestedAttribute {
	req, opt, comp := attributeModes(required, computed)
	return schema.ListNestedAttribute{
		MarkdownDescription: desc,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
		Required: req,
		Optional: opt,
		Computed: comp,
	}
}

func mapNestedAttr(desc string, required, computed bool, attrs map[string]schema.Attribute) schema.MapNestedAttribute {
	req, opt, comp := attributeModes(required, computed)
	return schema.MapNestedAttribute{
		MarkdownDescription: desc,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
		Required: req,
		Optional: opt,
		Computed: comp,
	}
}

func singleNestedAttr(desc string, required, computed bool, attrs map[string]schema.Attribute) schema.SingleNestedAttribute {
	req, opt, comp := attributeModes(required, computed)
	return schema.SingleNestedAttribute{
		MarkdownDescription: desc,
		Attributes:          attrs,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

func attributeModes(required, computed bool) (req, opt, comp bool) {
	switch {
	case required:
		return true, false, false
	case computed:
		return false, false, true
	default:
		return false, true, false
	}
}

func resiliencySpecAttribute(required, computed bool) schema.SingleNestedAttribute {
	return singleNestedAttr(
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
	return singleNestedAttr(
		"Resiliency policy definitions",
		false,
		computed,
		map[string]schema.Attribute{
			"timeouts":         mapStringAttr("Timeout policies keyed by name", false, computed),
			"retries":          resiliencyRetryPoliciesAttribute(computed),
			"circuit_breakers": resiliencyCircuitBreakerPoliciesAttribute(computed),
		},
	)
}

func resiliencyRetryPoliciesAttribute(computed bool) schema.MapNestedAttribute {
	return mapNestedAttr(
		"Retry policies keyed by name",
		false,
		computed,
		map[string]schema.Attribute{
			"duration":     stringAttr("Delay between retries (e.g. 5s)", false, computed),
			"max_interval": stringAttr("Maximum backoff interval", false, computed),
			"max_retries":  int64Attr("Maximum retry attempts (-1 for infinite)", false, computed),
			"policy":       stringAttr("Retry policy type (constant or exponential)", false, computed),
		},
	)
}

func resiliencyCircuitBreakerPoliciesAttribute(computed bool) schema.MapNestedAttribute {
	return mapNestedAttr(
		"Circuit breaker policies keyed by name",
		false,
		computed,
		map[string]schema.Attribute{
			"interval":     stringAttr("Time window used to calculate statistics", false, computed),
			"max_requests": int64Attr("Maximum requests allowed in half-open state", false, computed),
			"timeout":      stringAttr("Duration the circuit remains open", false, computed),
			"trip":         stringAttr("Condition that opens the circuit", false, computed),
		},
	)
}

func resiliencyTargetsAttribute(computed bool) schema.SingleNestedAttribute {
	return singleNestedAttr(
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
	return mapNestedAttr(
		"Application policy bindings keyed by app ID",
		false,
		computed,
		map[string]schema.Attribute{
			"circuit_breaker":            stringAttr("Circuit breaker policy name", false, computed),
			"circuit_breaker_cache_size": int64Attr("Size of the circuit breaker cache", false, computed),
			"retry":                      stringAttr("Retry policy name", false, computed),
			"timeout":                    stringAttr("Timeout policy name", false, computed),
		},
	)
}

func resiliencyTargetActorsAttribute(computed bool) schema.MapNestedAttribute {
	return mapNestedAttr(
		"Actor policy bindings keyed by actor type",
		false,
		computed,
		map[string]schema.Attribute{
			"circuit_breaker":            stringAttr("Circuit breaker policy name", false, computed),
			"circuit_breaker_cache_size": int64Attr("Size of the circuit breaker cache", false, computed),
			"circuit_breaker_scope":      stringAttr("Scope used for the circuit breaker", false, computed),
			"retry":                      stringAttr("Retry policy name", false, computed),
			"timeout":                    stringAttr("Timeout policy name", false, computed),
		},
	)
}

func resiliencyTargetComponentsAttribute(computed bool) schema.MapNestedAttribute {
	return mapNestedAttr(
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
	return singleNestedAttr(
		desc,
		false,
		computed,
		map[string]schema.Attribute{
			"circuit_breaker": stringAttr("Circuit breaker policy name", false, computed),
			"retry":           stringAttr("Retry policy name", false, computed),
			"timeout":         stringAttr("Timeout policy name", false, computed),
		},
	)
}
