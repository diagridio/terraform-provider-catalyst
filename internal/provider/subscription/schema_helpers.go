package subscription

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

func subscriptionSpecAttribute(required, computed bool) schema.SingleNestedAttribute {
	return singleNestedAttr(
		"Dapr subscription spec",
		required,
		computed,
		map[string]schema.Attribute{
			"routes":            subscriptionRoutesAttribute(computed),
			"bulk_subscribe":    subscriptionBulkSubscribeAttribute(computed),
			"dead_letter_topic": stringAttr("Dead letter topic", false, computed),
			"metadata":          mapStringAttr("Metadata entries", false, computed),
			"dynamic":           boolAttr("Dynamic subscription", false, computed),
		},
	)
}

func subscriptionRoutesAttribute(computed bool) schema.SingleNestedAttribute {
	return singleNestedAttr(
		"Routes configuration",
		false,
		computed,
		map[string]schema.Attribute{
			"default": stringAttr("Default route path", false, computed),
			"rules":   subscriptionRouteRulesAttribute(computed),
		},
	)
}

func subscriptionRouteRulesAttribute(computed bool) schema.ListNestedAttribute {
	return listNestedAttr(
		"Routing rules",
		false,
		computed,
		map[string]schema.Attribute{
			"match": stringAttr("Match expression", false, computed),
			"path":  stringAttr("Route path", false, computed),
		},
	)
}

func subscriptionBulkSubscribeAttribute(computed bool) schema.SingleNestedAttribute {
	return singleNestedAttr(
		"Bulk subscribe configuration",
		false,
		computed,
		map[string]schema.Attribute{
			"enabled":               boolAttr("Enable bulk subscribe", false, computed),
			"max_messages_count":    int64Attr("Maximum messages count", false, computed),
			"max_await_duration_ms": int64Attr("Maximum await duration in milliseconds", false, computed),
		},
	)
}
