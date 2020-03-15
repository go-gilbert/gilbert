package manifest

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

var anyType = cty.NilType

// getTypeAttrValue returns type name value of attribute by name
//
// see scalarTypeFromString for more information about supported values
func getTypeAttrValue(body *hclsyntax.Body, name string) (cty.Type, hcl.Diagnostics) {
	strType, diags := getStringAttrValue(body, name, nil, true)
	if diags != nil {
		return cty.NilType, diags
	}

	typ, err := scalarTypeFromString(strType)
	if err != nil {
		return cty.NilType, NewDiagnosticsFromPosition(
			body.Range(), "invalid type value in %q parameter: %s", name, err.Error(),
		)
	}

	return typ, nil
}

// getStringAttrValue returns string value of attribute by name
//
// if context is not nil, inner template expression will be evaluated
func getStringAttrValue(body *hclsyntax.Body, name string, ctx *hcl.EvalContext, required bool) (string, hcl.Diagnostics) {
	attr, ok := body.Attributes[name]
	if !ok {
		if !required {
			return "", nil
		}

		return "", NewDiagnosticsFromPosition(
			body.Range(), "%q parameter is required", name,
		)
	}

	var (
		val   cty.Value
		diags hcl.Diagnostics
	)

	if ctx != nil {
		val, diags = attr.Expr.Value(ctx)
	} else {
		val, diags = getLiteralAttrValue(attr, cty.String)
	}
	if diags != nil {
		return "", diags
	}

	if val.Type() != cty.String {
		return "", NewDiagnosticsFromPosition(
			attr.Range(), "invalid %q parameter value type(expected %q but got %q)",
			name, cty.String.FriendlyName(), val.Type().FriendlyName(),
		)
	}

	strVal := val.AsString()
	if strVal == "" {
		return "", NewDiagnosticsFromPosition(
			attr.Range(), "%q parameter is required", name,
		)
	}

	return strVal, nil
}

// getBoolAttrValue returns boolean value of attribute by name
func getBoolAttrValue(body *hclsyntax.Body, name string, otherwise bool) (bool, hcl.Diagnostics) {
	attr, ok := body.Attributes[name]
	if !ok {
		return otherwise, nil
	}

	val, diags := getLiteralAttrValue(attr, cty.Bool)
	if diags != nil {
		return false, diags
	}

	return val.True(), nil
}

// getAttrValue returns body attribute value by name and checks value type
//
// use manifest.anyType to omit type check
func getAttrValue(body *hclsyntax.Body, ctx *hcl.EvalContext, name string, expectType cty.Type, required bool) (cty.Value, hcl.Diagnostics) {
	attr, ok := body.Attributes[name]
	if !ok {
		if !required {
			return cty.NilVal, nil
		}

		return cty.NilVal, NewDiagnosticsFromPosition(
			body.Range(), "%q parameter is required", name,
		)
	}

	val, diags := attr.Expr.Value(ctx)
	if diags != nil {
		return cty.NilVal, diags
	}

	if expectType == anyType {
		// omit type check if expected type is any type
		return val, nil
	}

	if t := val.Type(); t != expectType {
		return cty.NilVal, NewDiagnosticsFromPosition(
			attr.Range(), "invalid %q parameter value. Expected %q but got %q",
			name, expectType.FriendlyName(), t.FriendlyName(),
		)
	}

	return val, nil
}

// getLiteralAttrValue checks if attribute value is expected type and returns it
func getLiteralAttrValue(attr *hclsyntax.Attribute, expectType cty.Type) (cty.Value, hcl.Diagnostics) {
	// attribute value should be literal
	expr, ok := attr.Expr.(*hclsyntax.LiteralValueExpr)
	if !ok {
		return cty.NilVal, NewDiagnosticsFromPosition(
			attr.Expr.StartRange(),
			"%q parameter attribute should be literal %s value",
			attr.Name, expectType.FriendlyName(),
		)
	}

	// check "required" attribute type and get value
	if valType := expr.Val.Type(); valType != cty.Bool {
		return cty.NilVal, NewDiagnosticsFromPosition(
			attr.Expr.Range(),
			"%q parameter attribute should be %s (got %q)",
			attr.Name, expectType.FriendlyName(), valType.FriendlyName(),
		)
	}

	return expr.Val, nil
}

// typeFromString returns go-cty type from string type name
func typeFromString(typeStr string, scalarOnly bool) (cty.Type, error) {
	typeStr = strings.TrimSpace(typeStr)
	if len(typeStr) == 0 {
		return cty.NilType, fmt.Errorf("type cannot be empty")
	}

	// Array type detection (like []string)
	if strings.HasPrefix(typeStr, arrayTypePrefix) {
		if scalarOnly {
			return cty.NilType, fmt.Errorf("only scalar types allowed (int, string, bool)")
		}
		if typeStr == arrayTypePrefix {
			return cty.NilType, fmt.Errorf("array type required")
		}

		arrayTypeStr := typeStr[len(arrayTypePrefix):]
		arrayItemType, err := typeFromString(arrayTypeStr, false)
		if err != nil {
			return cty.NilType, fmt.Errorf("invalid array type %q - %w", typeStr, err)
		}

		return cty.List(arrayItemType), nil
	}

	return scalarTypeFromString(typeStr)
}

// scalarTypeFromString returns scalar go-cty type by string name
func scalarTypeFromString(typeStr string) (cty.Type, error) {
	switch typeStr {
	case "string":
		return cty.String, nil
	case "number":
		return cty.Number, nil
	case "bool":
		return cty.Bool, nil
	default:
		return cty.NilType, fmt.Errorf("unknown type %q", typeStr)
	}
}
