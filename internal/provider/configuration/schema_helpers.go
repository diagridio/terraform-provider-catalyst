package configuration

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringAttr(desc string, required, computed bool) schema.StringAttribute {
	attr := schema.StringAttribute{MarkdownDescription: desc}
	switch {
	case required:
		attr.Required = true
	case computed:
		attr.Computed = true
	default:
		attr.Optional = true
	}
	return attr
}

func stringAttrOptionalComputed(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
	}
}

func stringAttrComputed(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Computed:            true,
	}
}

func boolAttr(desc string, required, computed bool) schema.BoolAttribute {
	attr := schema.BoolAttribute{MarkdownDescription: desc}
	switch {
	case required:
		attr.Required = true
	case computed:
		attr.Computed = true
	default:
		attr.Optional = true
	}
	return attr
}

func listAttr(desc string, elemType attr.Type, required, computed bool) schema.ListAttribute {
	attr := schema.ListAttribute{
		MarkdownDescription: desc,
		ElementType:         elemType,
	}
	switch {
	case required:
		attr.Required = true
	case computed:
		attr.Computed = true
	default:
		attr.Optional = true
	}
	return attr
}

func mapAttr(desc string, elemType attr.Type, required, computed bool) schema.MapAttribute {
	attr := schema.MapAttribute{
		MarkdownDescription: desc,
		ElementType:         elemType,
	}
	switch {
	case required:
		attr.Required = true
	case computed:
		attr.Computed = true
	default:
		attr.Optional = true
	}
	return attr
}

func singleNestedAttr(desc string, required, computed bool, attrs map[string]schema.Attribute) schema.SingleNestedAttribute {
	attr := schema.SingleNestedAttribute{
		MarkdownDescription: desc,
		Attributes:          attrs,
	}
	switch {
	case required:
		attr.Required = true
	case computed:
		attr.Computed = true
	default:
		attr.Optional = true
	}
	return attr
}

func listNestedAttr(desc string, required, computed bool, attrs map[string]schema.Attribute) schema.ListNestedAttribute {
	attr := schema.ListNestedAttribute{
		MarkdownDescription: desc,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
	}
	switch {
	case required:
		attr.Required = true
	case computed:
		attr.Computed = true
	default:
		attr.Optional = true
	}
	return attr
}

func configurationSpecAttribute(required, computed bool) schema.SingleNestedAttribute {
	return singleNestedAttr(
		"Dapr configuration spec",
		required,
		computed,
		map[string]schema.Attribute{
			"access_control": singleNestedAttr(
				"Access control configuration",
				false,
				computed,
				map[string]schema.Attribute{
					"default_action": stringAttr("Default action when no policy matches", false, computed),
					"trust_domain":   stringAttr("Trust domain for access control", false, computed),
					"policies": listNestedAttr(
						"Access control policies",
						false,
						computed,
						map[string]schema.Attribute{
							"app_id":         stringAttr("Application ID this policy applies to", false, computed),
							"default_action": stringAttr("Default action for the policy", false, computed),
							"namespace":      stringAttrComputed("Namespace constraint for the policy (computed by API)"),
							"trust_domain":   stringAttr("Policy-specific trust domain", false, computed),
							"operations": listNestedAttr(
								"Operations governed by the policy",
								false,
								computed,
								map[string]schema.Attribute{
									"name":       stringAttr("Operation name", false, computed),
									"action":     stringAttr("Action applied to the operation", false, computed),
									"http_verbs": listAttr("Allowed HTTP verbs", types.StringType, false, computed),
								},
							),
						},
					),
				},
			),
			"app_http_pipeline": singleNestedAttr(
				"Application HTTP pipeline handlers",
				false,
				computed,
				map[string]schema.Attribute{
					"handlers": listNestedAttr(
						"Handlers executed for app HTTP traffic",
						false,
						computed,
						map[string]schema.Attribute{
							"name": stringAttr("Handler name", false, computed),
							"type": stringAttr("Handler type", false, computed),
						},
					),
				},
			),
			"http_pipeline": singleNestedAttr(
				"Global HTTP pipeline handlers",
				false,
				computed,
				map[string]schema.Attribute{
					"handlers": listNestedAttr(
						"Handlers executed for ingress HTTP traffic",
						false,
						computed,
						map[string]schema.Attribute{
							"name": stringAttr("Handler name", false, computed),
							"type": stringAttr("Handler type", false, computed),
						},
					),
				},
			),
		},
	)
}
