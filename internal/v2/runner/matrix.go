package runner

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type (
	LazyMatrix = map[string]*manifest.LazyValue
	Matrix     = map[string][]any
)

type ExecutionMatrix struct {
	Labels []string
	Values [][]any
}

type matrixParam struct {
	key      string
	variants []any
}

func resolveMatrixValues(ctx context.Context, ep expr.EvalParams, mat []manifest.MatrixParam) ([]matrixParam, parsetypes.Diagnostics) {
	var diags parsetypes.Diagnostics
	out := make([]matrixParam, 0, len(mat))
	for _, mp := range mat {
		lval := mp.Values
		if lval.Value.ArraySpec != nil && lval.Value.ArraySpec.Literal() {
			// micro-op when got literal value
			out = append(out, matrixParam{
				key:      mp.Key,
				variants: lval.Value.ArraySpec.LiteralItems,
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
		return nil, fmt.Errorf("matrix values should be a list, got %v", a)
	}

	ref := reflect.ValueOf(a)
	switch k := ref.Kind(); k {
	case reflect.Slice, reflect.Array:
		break
	default:
		return nil, fmt.Errorf("matrix values should be a list, got %s %#v", k, a)
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
