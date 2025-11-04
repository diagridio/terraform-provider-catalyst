package configuration

import (
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/helpers"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func configurationSpecAttribute(required, computed bool) schema.SingleNestedAttribute {
	return helpers.SingleNestedAttr(
		"Dapr configuration spec",
		required,
		computed,
		map[string]schema.Attribute{
			"access_control": helpers.SingleNestedAttr(
				"Access control configuration",
				false,
				computed,
				map[string]schema.Attribute{
					"default_action": helpers.StringAttr("Default action when no policy matches", computed),
					"trust_domain":   helpers.StringAttr("Trust domain for access control", computed),
					"policies": helpers.ListNestedAttr(
						"Access control policies",
						false,
						computed,
						map[string]schema.Attribute{
							"app_id":         helpers.StringAttr("Application ID this policy applies to", computed),
							"default_action": helpers.StringAttr("Default action for the policy", computed),
							"namespace":      helpers.StringAttrComputed("Namespace constraint for the policy (computed by API)"),
							"trust_domain":   helpers.StringAttr("Policy-specific trust domain", computed),
							"operations": helpers.ListNestedAttr(
								"Operations governed by the policy",
								false,
								computed,
								map[string]schema.Attribute{
									"name":       helpers.StringAttr("Operation name", computed),
									"action":     helpers.StringAttr("Action applied to the operation", computed),
									"http_verbs": helpers.ListAttr("Allowed HTTP verbs", types.StringType, false, computed),
								},
							),
						},
					),
				},
			),
			"app_http_pipeline": helpers.SingleNestedAttr(
				"Application HTTP pipeline handlers",
				false,
				computed,
				map[string]schema.Attribute{
					"handlers": helpers.ListNestedAttr(
						"Handlers executed for app HTTP traffic",
						false,
						computed,
						map[string]schema.Attribute{
							"name": helpers.StringAttr("Handler name", computed),
							"type": helpers.StringAttr("Handler type", computed),
						},
					),
				},
			),
			"http_pipeline": helpers.SingleNestedAttr(
				"Global HTTP pipeline handlers",
				false,
				computed,
				map[string]schema.Attribute{
					"handlers": helpers.ListNestedAttr(
						"Handlers executed for ingress HTTP traffic",
						false,
						computed,
						map[string]schema.Attribute{
							"name": helpers.StringAttr("Handler name", computed),
							"type": helpers.StringAttr("Handler type", computed),
						},
					),
				},
			),
		},
	)
}
