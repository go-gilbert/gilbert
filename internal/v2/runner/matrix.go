package runner

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"reflect"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type ExecutionMatrix struct {
	Labels []string
	Values [][]any
}

// testMatrixExcluded checks whether matrix values are matching any exclude rules.
//
// Method assumes that matrix value types already checked by [validateMatrixExcludeRules] for comparability.
func testMatrixExcluded(es *manifest.ExecStrategy, values []any) bool {
	if len(es.Exclude) == 0 {
		return false
	}

	for _, rule := range es.Exclude {
		matchLeft := len(rule)
		for k, v := range rule {
			j, ok := es.MatrixKeys[k]
			if !ok {
				continue
			}

			// exclude rule values already typechecked by yamlloader.
			// matrix parameter value already checked by validateMatrixExcludeRules.
			ruleVal := v.Value
			gotVal := values[j]

			if gotVal == ruleVal {
				matchLeft--
			}
		}

		if matchLeft == 0 {
			return true
		}
	}

	return false
}

type matrixParam struct {
	key      string
	variants []any
}

// validateMatrixExcludeRules checks whether any of exclude rules are referencing matrix parameters with non-scalar values.
func validateMatrixExcludeRules(es *manifest.ExecStrategy, matValues []matrixParam) parsetypes.Diagnostics {
	if len(es.Exclude) == 0 {
		return nil
	}

	alreadyCheckedKeys := make(map[string]bool, len(matValues))
	var diags parsetypes.Diagnostics
	for _, rule := range es.Exclude {
		for k, v := range rule {
			// exclude rule values already typechecked by yamlloader.
			ruleVal := v.Value
			if !parsetypes.IsScalar(ruleVal) {
				continue
			}

			if isParamValid, ok := alreadyCheckedKeys[k]; ok {
				// Skip check of already checked matrix params.
				if !isParamValid {
					diags = append(diags, &parsetypes.Diagnostic{
						FileName: v.Location.FileName,
						Severity: parsetypes.DiagnosticSeverityError,
						Range:    v.Location.Range,
						Offset:   v.Location.Offset,
						Err:      fmt.Errorf("exclude rule references a non-comparable matrix parameter %q", k),
					})
				}

				continue
			}

			ii, ok := es.MatrixKeys[k]
			if !ok {
				diags = append(diags, &parsetypes.Diagnostic{
					FileName: v.Location.FileName,
					Severity: parsetypes.DiagnosticSeverityError,
					Range:    v.Location.Range,
					Offset:   v.Location.Offset,
					Note:     fmt.Sprintf("matrix doesn't contain a value with key %q", k),
					Err:      fmt.Errorf("exclude rule references a non-existing matrix parameter %q", k),
				})
				continue
			}

			// Check if all slice items in matrix parameter are scalar and thus, comparable.
			isParamValid := true
			for j, v := range matValues[ii].variants {
				if parsetypes.IsScalar(v) {
					continue
				}

				// mark param as invalid
				isParamValid = false
				spec := es.Matrix[ii].Values
				if spec.Location == nil {
					continue
				}

				diags = append(diags, &parsetypes.Diagnostic{
					FileName: spec.Location.FileName,
					Severity: parsetypes.DiagnosticSeverityError,
					Range:    spec.Location.Range,
					Offset:   spec.Location.Offset,
					Note:     fmt.Sprintf("value %s at index %d is not comparable", parsetypes.Spew(v), j),
					Err:      fmt.Errorf("matrix parameter %q contains uncomparable value and cannot be used in exclude rule", k),
				})
				break
			}

			alreadyCheckedKeys[k] = isParamValid
			if !isParamValid {
				diags = append(diags, &parsetypes.Diagnostic{
					FileName: v.Location.FileName,
					Severity: parsetypes.DiagnosticSeverityError,
					Range:    v.Location.Range,
					Offset:   v.Location.Offset,
					Err:      fmt.Errorf("exclude rule references a non-comparable matrix parameter %q", k),
				})
			}
		}
	}

	return diags
}

// innerJoinMatrix returns an iterator that iterates over a cartesian product of matrix values.
func innerJoinMatrix(mps []matrixParam) iter.Seq2[[]string, []any] {
	// Do catersian product of all possible combination of matrix values
	return func(yield func([]string, []any) bool) {
		keys := make([]string, len(mps))
		totalRows := 1
		for i, v := range mps {
			keys[i] = v.key
			totalRows *= len(v.variants)
		}

		for i := range totalRows {
			row := make([]any, len(mps))
			idx := i
			for j := len(mps) - 1; j >= 0; j-- {
				varcount := len(mps[j].variants)
				row[j] = mps[j].variants[idx%varcount]
				idx /= varcount
			}

			if !yield(keys, row) {
				return
			}
		}
	}
}

func validateMatrixParam(spec *manifest.LazyValue, key string, vals []any) (bool, parsetypes.Diagnostics) {
	var diags parsetypes.Diagnostics
	if len(vals) == 0 {
		var note string
		if !spec.Value.IsLiteral() {
			note = "expression returned an empty value"
		}

		diags = append(diags, &parsetypes.Diagnostic{
			FileName: spec.Location.FileName,
			Severity: parsetypes.DiagnosticSeverityWarning,
			Range:    spec.Location.Range,
			Offset:   spec.Location.Offset,
			Note:     note,
			Err:      fmt.Errorf("matrix parameter %q has no values and will be ignored", key),
		})
		return false, diags
	}

	return true, nil
}

func resolveMatrixValues(ctx context.Context, ep expr.EvalParams, mat []manifest.MatrixParam) ([]matrixParam, parsetypes.Diagnostics) {
	var diags parsetypes.Diagnostics
	out := make([]matrixParam, 0, len(mat))
	for _, mp := range mat {
		lval := mp.Values
		if lval.Value.ArraySpec != nil && lval.Value.ArraySpec.Literal() {
			// micro-op when got literal value
			litItems := lval.Value.ArraySpec.LiteralItems
			ok, diag := validateMatrixParam(lval, mp.Key, litItems)
			diags = append(diags, diag...)
			if !ok {
				continue
			}

			out = append(out, matrixParam{
				key:      mp.Key,
				variants: litItems,
			})
			continue
		}

		// Stage 1: resolve thunk
		anyVal, err := lval.Expand(ctx, ep)
		if err != nil {
			if d, ok := parsetypes.IsDiagnosticsError(err); ok {
				diags = append(diags, d...)
				continue
			}

			diags = append(diags, &parsetypes.Diagnostic{
				FileName: lval.Location.FileName,
				Severity: parsetypes.DiagnosticSeverityError,
				Range:    lval.Location.Range,
				Offset:   lval.Location.Offset,
				Note:     err.Error(),
				Err:      fmt.Errorf("failed to resolve values for matrix parameter %q", mp.Key),
			})
			continue
		}

		// Stage 2: type check and transform into a slice
		// Slice item type is checked later and only if used in exluce rules.
		vals, err := anyToArray(anyVal)
		if err != nil {
			diags = append(diags, &parsetypes.Diagnostic{
				FileName: lval.Location.FileName,
				Severity: parsetypes.DiagnosticSeverityError,
				Range:    lval.Location.Range,
				Offset:   lval.Location.Offset,
				Note:     err.Error(),
				Err:      fmt.Errorf("invalid value type for matrix parameter %q", mp.Key),
			})
			continue
		}

		// Sanify check - empty values
		ok, diag := validateMatrixParam(lval, mp.Key, vals)
		diags = append(diags, diag...)
		if !ok {
			continue
		}

		out = append(out, matrixParam{
			key:      mp.Key,
			variants: vals,
		})
	}

	return out, diags
}

func anyToArray(a any) (out []any, err error) {
	defer func() {
		// reflect is kinda panicky
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if v, ok := a.([]any); ok {
		return v, nil
	}

	if a == nil {
		return nil, fmt.Errorf("expected a list, but got %v", a)
	}

	ref := reflect.ValueOf(a)
	switch k := ref.Kind(); k {
	case reflect.Slice, reflect.Array:
		break
	default:
		return nil, fmt.Errorf("expected a list, but got %s %#v", k, a)
	}

	arrLen := ref.Len()
	if arrLen == 0 {
		return nil, errors.New("matrix values list cannot be empty")
	}

	out = make([]any, arrLen)
	for i := range arrLen {
		out[i] = ref.Index(i)
	}

	return out, nil
}
