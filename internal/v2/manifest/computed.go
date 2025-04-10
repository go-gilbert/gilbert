package manifest

import (
	"context"

	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type ReferenceLocation struct {
	FileName string
	Range    parsetypes.Range
	Offset   parsetypes.OffsetRange
}

type TypedLazyValue struct {
	Type  ValueType
	Value LazyValue
}

func (tlz TypedLazyValue) Expand(ctx context.Context, opts expr.EvalParams) (any, error) {
	// TODO: typecheck?
	return tlz.Value.Expand(ctx, opts)
}

// IsType checks if lazy value type matches to a schema
func (tlz TypedLazyValue) IsType(t TypeSchema) bool {
	return tlz.Type == t.Type
}

// LazyValue represents a computed value that contains template expression
// that needs to be evaluated in runtime.
type LazyValue struct {
	Location *ReferenceLocation
	Value    AnySpec
}

func (v *LazyValue) Expand(ctx context.Context, opts expr.EvalParams) (any, error) {
	// TODO: convert into diagnostics
	return v.Value.Expand(ctx, opts)
}

type LiteralSpec struct {
	Value any
}

type BindingSpec struct {
	Location ReferenceLocation
	Expr     expr.Expression
}

type ArraySpec struct {
	LiteralItems []any
	DynamicItems []AnySpec
}

func (s ArraySpec) Literal() bool {
	return len(s.LiteralItems) > 0
}

func (s ArraySpec) Expand(ctx context.Context, opts expr.EvalParams) (any, error) {
	if len(s.LiteralItems) != 0 {
		return s.LiteralItems, nil
	}

	dst := make([]any, len(s.DynamicItems))
	for i, spec := range s.DynamicItems {
		val, err := spec.Expand(ctx, opts)
		if err != nil {
			return nil, err
		}

		dst[i] = val
	}

	return dst, nil
}

func (s ArraySpec) Optimize() AnySpec {
	if s.Literal() {
		return AnySpec{
			LiteralSpec: &LiteralSpec{
				Value: s.LiteralItems,
			},
		}
	}

	hasBinding := false
	for i, item := range s.DynamicItems {
		opt := item.Optimize()
		s.DynamicItems[i] = opt

		if !opt.IsLiteral() {
			hasBinding = true
		}
	}

	if hasBinding {
		return AnySpec{
			ArraySpec: &s,
		}
	}

	literals := make([]any, len(s.DynamicItems))
	for i, v := range s.DynamicItems {
		literals[i] = v.LiteralSpec.Value
	}

	return AnySpec{
		LiteralSpec: &LiteralSpec{
			Value: literals,
		},
	}
}

type AnySpec struct {
	ArraySpec   *ArraySpec
	ObjectSpec  *ObjectSpec
	BindingSpec *BindingSpec
	LiteralSpec *LiteralSpec
}

func (s AnySpec) IsLiteral() bool {
	return s.LiteralSpec != nil
}

func (s AnySpec) Optimize() AnySpec {
	if s.LiteralSpec != nil {
		return s
	}

	if s.ArraySpec != nil {
		return s.ArraySpec.Optimize()
	}

	if s.ObjectSpec != nil {
		return s.ObjectSpec.Optimize()
	}

	return s
}

func (s AnySpec) Expand(ctx context.Context, opts expr.EvalParams) (any, error) {
	if s.LiteralSpec != nil {
		return s.LiteralSpec.Value, nil
	}

	if s.BindingSpec != nil {
		// TODO: convert into diagnostics
		return s.BindingSpec.Expr.Eval(ctx, opts)
	}

	if s.ObjectSpec != nil {
		return s.ObjectSpec.Expand(ctx, opts)
	}

	if s.ObjectSpec != nil {
		return s.ObjectSpec.Expand(ctx, opts)
	}

	if s.ObjectSpec != nil {
		return s.ObjectSpec.Expand(ctx, opts)
	}

	return nil, nil
}

type ObjectSpec struct {
	Values map[string]AnySpec
}

func (s ObjectSpec) Expand(ctx context.Context, opts expr.EvalParams) (any, error) {
	dst := make(map[string]any, len(s.Values))

	for k, spec := range s.Values {
		val, err := spec.Expand(ctx, opts)
		if err != nil {
			return nil, err
		}

		dst[k] = val
	}

	return dst, nil
}

func (s ObjectSpec) Optimize() AnySpec {
	hasBinding := false
	for k, item := range s.Values {
		opt := item.Optimize()
		if opt.IsLiteral() {
			hasBinding = true
		}

		s.Values[k] = opt
	}

	if hasBinding {
		return AnySpec{
			ObjectSpec: &s,
		}
	}

	out := make(map[string]any, len(s.Values))
	for k, v := range s.Values {
		out[k] = v.LiteralSpec.Value
	}

	return AnySpec{
		LiteralSpec: &LiteralSpec{
			Value: out,
		},
	}
}
