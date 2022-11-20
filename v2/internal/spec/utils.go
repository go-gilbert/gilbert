package spec

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"strings"
)

func lookupLabelAndDescription(labels []string) (name, desc string) {
	return lookupLabel(labels, 0), lookupLabel(labels, 1)
}

func lookupLabel(labels []string, index int) string {
	if len(labels) > index {
		return strings.TrimSpace(labels[index])
	}

	return ""
}

func extractListAttr[T any](name string, body *hclsyntax.Body, ctx *hcl.EvalContext) ([]T, hcl.Diagnostics) {
	attr, ok := body.Attributes[name]
	if !ok {
		return nil, nil
	}

	val, err := attr.Expr.Value(ctx)
	if err != nil {
		return nil, newDiagnosticError(attr.Range(), "failed to parse %q attribute: %s",
			attr.Name, err)
	}

	if val == cty.NilVal {
		return nil, err
	}

	out, cpErr := ctyTupleToSlice[T](val)
	if cpErr != nil {
		return nil, newDiagnosticError(attr.Range(),
			"invalid %q attribute value: %s", attr.Name, cpErr)
	}

	return out, nil
}

func extractAttr[T any](name string, body *hclsyntax.Body, ctx *hcl.EvalContext) (T, bool, hcl.Diagnostics) {
	var out T
	attr, ok := body.Attributes[name]
	if !ok {
		return out, false, nil
	}

	val, err := attr.Expr.Value(ctx)
	if err != nil {
		return out, true, newDiagnosticError(attr.Range(), "failed to parse %q attribute: %s",
			attr.Name, err)
	}

	if val == cty.NilVal {
		return out, false, nil
	}

	if err := gocty.FromCtyValue(val, &out); err != nil {
		return out, true, newDiagnosticError(attr.Range(),
			"invalid %q attribute value: %s", attr.Name, err)
	}

	return out, true, nil
}

func unmarshalAttr[T any](attr *hcl.Attribute, ctx *hcl.EvalContext) (T, hcl.Diagnostics) {
	var out T
	val, err := attr.Expr.Value(ctx)
	if err != nil {
		return out, newDiagnosticError(attr.Range,
			"failed to parse %q attribute: %s", attr.Name, err)
	}

	if err := gocty.FromCtyValue(val, &out); err != nil {
		return out, newDiagnosticError(attr.Range,
			"invalid %q attribute value: %s", attr.Name, err)
	}

	return out, nil
}

func ctyTupleToSlice[T any](val cty.Value) ([]T, error) {
	var out []T
	if !val.Type().IsTupleType() {
		err := gocty.FromCtyValue(val, out)
		return out, err
	}

	out = make([]T, val.Type().Length())
	for i := range out {
		elem := val.Index(cty.NumberIntVal(int64(i)))
		if err := gocty.FromCtyValue(elem, &out[i]); err != nil {
			return nil, fmt.Errorf("invalid element at index %d, %w", i, err)
		}
	}

	return out, nil
}

func newDiagnosticError(rng hcl.Range, msg string, args ...any) hcl.Diagnostics {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     msg,
			Subject:     &rng,
			Context:     &rng,
			Expression:  nil,
			EvalContext: nil,
			Extra:       nil,
		},
	}
}
