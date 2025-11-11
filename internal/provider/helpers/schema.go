package helpers

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringAttr creates a string attribute with markdown description and computed/optional mode.
func StringAttr(desc string, computed bool) schema.StringAttribute {
	req, opt, comp := AttributeModes(false, computed)
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

// StringAttrComputed creates a computed-only string attribute with markdown description.
func StringAttrComputed(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Computed:            true,
	}
}

// BoolAttr creates a boolean attribute with markdown description and required/computed mode.
func BoolAttr(desc string, required, computed bool) schema.BoolAttribute {
	req, opt, comp := AttributeModes(required, computed)
	return schema.BoolAttribute{
		MarkdownDescription: desc,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

// Int64Attr creates an int64 attribute with markdown description and required/computed mode.
func Int64Attr(desc string, required, computed bool) schema.Int64Attribute {
	req, opt, comp := AttributeModes(required, computed)
	return schema.Int64Attribute{
		MarkdownDescription: desc,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

// MapStringAttr creates a map attribute with string element type, markdown description and required/computed mode.
func MapStringAttr(desc string, required, computed bool) schema.MapAttribute {
	req, opt, comp := AttributeModes(required, computed)
	return schema.MapAttribute{
		MarkdownDescription: desc,
		ElementType:         types.StringType,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

// ListAttr creates a list attribute with markdown description, element type and required/computed mode.
func ListAttr(desc string, elemType attr.Type, required, computed bool) schema.ListAttribute {
	req, opt, comp := AttributeModes(required, computed)
	return schema.ListAttribute{
		MarkdownDescription: desc,
		ElementType:         elemType,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

// ListNestedAttr creates a list nested attribute with markdown description and required/computed mode.
func ListNestedAttr(desc string, required, computed bool, attrs map[string]schema.Attribute) schema.ListNestedAttribute {
	req, opt, comp := AttributeModes(required, computed)
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

// SingleNestedAttr creates a single nested attribute with markdown description and required/computed mode.
func SingleNestedAttr(desc string, required, computed bool, attrs map[string]schema.Attribute) schema.SingleNestedAttribute {
	req, opt, comp := AttributeModes(required, computed)
	return schema.SingleNestedAttribute{
		MarkdownDescription: desc,
		Attributes:          attrs,
		Required:            req,
		Optional:            opt,
		Computed:            comp,
	}
}

// MapNestedAttr creates a map nested attribute with markdown description and required/computed mode.
func MapNestedAttr(desc string, required, computed bool, attrs map[string]schema.Attribute) schema.MapNestedAttribute {
	req, opt, comp := AttributeModes(required, computed)
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

// AttributeModes returns the required, optional, and computed flags based on the input parameters.
func AttributeModes(required, computed bool) (req, opt, comp bool) {
	switch {
	case required:
		return true, false, false
	case computed:
		return false, false, true
	default:
		return false, true, false
	}
}
