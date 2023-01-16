package ctyutil

import (
	"strconv"

	"github.com/zclconf/go-cty/cty"
)

// ValueToString converts cty.Value to string representation.
//
// Supports only primitive values.
func ValueToString(val cty.Value) string {
	if val == cty.NilVal || val.IsNull() {
		return ""
	}

	switch typ := val.Type(); typ {
	case cty.Bool:
		return strconv.FormatBool(val.True())
	case cty.Number:
		return val.AsBigFloat().String()
	case cty.String:
		return val.AsString()
	default:
		return val.GoString()
	}
}
