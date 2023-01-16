package spec

import (
	"fmt"
	"github.com/go-gilbert/gilbert/v2/internal/util/hclutil"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func extractListAttr[T any](name string, attrs hclsyntax.Attributes, ctx *hcl.EvalContext) ([]T, hcl.Diagnostics) {
	attr, ok := attrs[name]
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

func extractAttr[T any](name string, attrs hclsyntax.Attributes, ctx *hcl.EvalContext) (T, bool, hcl.Diagnostics) {
	var out T
	attr, ok := attrs[name]
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
	return hclutil.NewDiagnosticError(rng, msg, args...)
}
