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

type matrixParam struct {
	key      string
	variants []any
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

func resolveMatrixValues(ctx context.Context, ep expr.EvalParams, mat []manifest.MatrixParam) ([]matrixParam, parsetypes.Diagnostics) {
	var diags parsetypes.Diagnostics
	out := make([]matrixParam, 0, len(mat))
	for _, mp := range mat {
		lval := mp.Values
		if lval.Value.ArraySpec != nil && lval.Value.ArraySpec.Literal() {
			// micro-op when got literal value
			litItems := lval.Value.ArraySpec.LiteralItems
			if len(litItems) == 0 {
				diags = append(diags, &parsetypes.Diagnostic{
					FileName: lval.Location.FileName,
					Severity: parsetypes.DiagnosticSeverityWarning,
					Range:    lval.Location.Range,
					Offset:   lval.Location.Offset,
					Note:     "empty array specified",
					Err:      fmt.Errorf("matrix parameter %q has no values and will be ignored", mp.Key),
				})
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
		if len(vals) == 0 {
			var note string
			if !lval.Value.IsLiteral() {
				note = "expression returned an empty value"
			}

			diags = append(diags, &parsetypes.Diagnostic{
				FileName: lval.Location.FileName,
				Severity: parsetypes.DiagnosticSeverityWarning,
				Range:    lval.Location.Range,
				Offset:   lval.Location.Offset,
				Note:     note,
				Err:      fmt.Errorf("matrix parameter %q has no values and will be ignored", mp.Key),
			})
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
