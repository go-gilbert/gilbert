package manifest2

import (
	"github.com/go-gilbert/gilbert/internal/manifest/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type ReferenceLocation struct {
	FileName string
	Range    parsetypes.Range
	Offset   parsetypes.OffsetRange
}

type TypedLazyValue struct {
	Type   ValueType
	Format ValueFormat
	Value  LazyValue
}

// LazyValue represents a computed value that contains template expression
// that needs to be evaluated in runtime.
type LazyValue struct {
	Location *ReferenceLocation
	Value    AnySpec
}

func (v *LazyValue) Expand(ctx expr.EvalContext) (any, error) {
	panic("not implemented")
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

type ObjectSpec struct {
	Values map[string]AnySpec
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
