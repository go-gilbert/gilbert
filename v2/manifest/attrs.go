package manifest

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type OrderedAttributes []*hclsyntax.Attribute

// OrderedAttributesFromBody creates slice of attributes sorted
// in declaration order from hcl body
//
// See: https://github.com/hashicorp/hcl/pull/352
func OrderedAttributesFromBody(b *hclsyntax.Body) OrderedAttributes {
	// this method available only in Gilbert's fork of hcl/v2
	// see: https://github.com/hashicorp/hcl/pull/352
	keys := b.AttributeNames()
	out := make(OrderedAttributes, len(keys))

	for i, key := range keys {
		out[i] = b.Attributes[key]
	}

	return out
}
