package configuration

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringAttr(desc string, computed bool) schema.StringAttribute {
	attr := schema.StringAttribute{MarkdownDescription: desc}
	if computed {
		attr.Computed = true
	} else {
		attr.Optional = true
	}
	return attr
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

func listNestedAttr(desc string, computed bool, attrs map[string]schema.Attribute) schema.ListNestedAttribute {
	attr := schema.ListNestedAttribute{
		MarkdownDescription: desc,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
	}
	if computed {
		attr.Computed = true
	} else {
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
					"default_action": stringAttr("Default action when no policy matches", computed),
					"trust_domain":   stringAttr("Trust domain for access control", computed),
					"policies": listNestedAttr(
						"Access control policies",
						computed,
						map[string]schema.Attribute{
							"app_id":         stringAttr("Application ID this policy applies to", computed),
							"default_action": stringAttr("Default action for the policy", computed),
							"namespace":      stringAttrComputed("Namespace constraint for the policy (computed by API)"),
							"trust_domain":   stringAttr("Policy-specific trust domain", computed),
							"operations": listNestedAttr(
								"Operations governed by the policy",
								computed,
								map[string]schema.Attribute{
									"name":       stringAttr("Operation name", computed),
									"action":     stringAttr("Action applied to the operation", computed),
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
						computed,
						map[string]schema.Attribute{
							"name": stringAttr("Handler name", computed),
							"type": stringAttr("Handler type", computed),
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
						computed,
						map[string]schema.Attribute{
							"name": stringAttr("Handler name", computed),
							"type": stringAttr("Handler type", computed),
						},
					),
				},
			),
		},
	)
}
