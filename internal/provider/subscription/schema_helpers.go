package subscription

import (
	"github.com/diagridio/terraform-provider-catalyst/internal/provider/helpers"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func subscriptionSpecAttribute(required, computed bool) schema.SingleNestedAttribute {
	return helpers.SingleNestedAttr(
		"Dapr subscription spec",
		required,
		computed,
		map[string]schema.Attribute{
			"routes":            subscriptionRoutesAttribute(computed),
			"bulk_subscribe":    subscriptionBulkSubscribeAttribute(computed),
			"dead_letter_topic": helpers.StringAttr("Dead letter topic", computed),
			"metadata":          helpers.MapStringAttr("Metadata entries", false, computed),
		},
	)
}

func subscriptionRoutesAttribute(computed bool) schema.SingleNestedAttribute {
	return helpers.SingleNestedAttr(
		"Routes configuration",
		false,
		computed,
		map[string]schema.Attribute{
			"default": helpers.StringAttr("Default route path", computed),
			"rules":   subscriptionRouteRulesAttribute(computed),
		},
	)
}

func subscriptionRouteRulesAttribute(computed bool) schema.ListNestedAttribute {
	return helpers.ListNestedAttr(
		"Routing rules",
		false,
		computed,
		map[string]schema.Attribute{
			"match": helpers.StringAttr("Match expression", computed),
			"path":  helpers.StringAttr("Route path", computed),
		},
	)
}

func subscriptionBulkSubscribeAttribute(computed bool) schema.SingleNestedAttribute {
	return helpers.SingleNestedAttr(
		"Bulk subscribe configuration",
		false,
		computed,
		map[string]schema.Attribute{
			"enabled":               helpers.BoolAttr("Enable bulk subscribe", false, computed),
			"max_messages_count":    helpers.Int64Attr("Maximum messages count", false, computed),
			"max_await_duration_ms": helpers.Int64Attr("Maximum await duration in milliseconds", false, computed),
		},
	)
}
