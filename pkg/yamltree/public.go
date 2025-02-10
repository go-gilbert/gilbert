// Package yamltree provides tools for unmarshaling and validation of YAML AST trees without reflection.
package yamltree

import (
	"context"
	"time"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml/ast"
	"golang.org/x/exp/constraints"
)

type UnknownFieldAction uint8

const (
	UnknownFieldActionIgnore UnknownFieldAction = iota
	UnknownFieldActionWarn
	UnknownFieldActionError
)

type TraverseOpts struct {
	FileName           string
	UnknownFieldAction UnknownFieldAction
}

// ValueVisitor is interface to implement decoding YAML AST nodes to values.
type ValueVisitor[T any] interface {
	VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) (T, parsetypes.Diagnostics)
}

// UInt returns decoder for unsigned integer values.
func UInt[T constraints.Unsigned]() ValueVisitor[T] {
	return uintVisitor[T]{}
}

// Int returns decoder for signed integer values.
func Int[T constraints.Signed]() ValueVisitor[T] {
	return intVisitor[T]{}
}

// Bool returns decoder for boolean values.
func Bool() ValueVisitor[bool] {
	return boolVisitor{}
}

// Float returns decoder for float values.
func Float[T constraints.Float]() ValueVisitor[T] {
	return floatVisitor[T]{}
}

// WithStrictStringValue option requires YAML node to be strictly string.
func WithStrictStringValue() func(*stringVisitor) {
	return func(v *stringVisitor) {
		v.strict = true
	}
}

// String returns decoder for strings.
//
// By default, decoder casts non-string primitives into a string.
// To disable this, use WithStrictStringValue() option.
func String(opts ...func(*stringVisitor)) ValueVisitor[string] {
	v := stringVisitor{}
	for _, opt := range opts {
		opt(&v)
	}

	return v
}

// Pointer wraps a decoder to return a pointer to a value.
func Pointer[T any](dec ValueVisitor[T]) ValueVisitor[*T] {
	return ptrVisitor[T]{
		visitor: dec,
	}
}

// List returns decoder for arrays.
func List[T any](itemDec ValueVisitor[T]) ValueVisitor[[]T] {
	return listVisitor[T]{
		listItemVisitor: itemDec,
	}
}

// Field returns decoder for object property.
//
// Meant to be passed inside Struct().
func Field[TObject, TProp any](
	propDec ValueVisitor[TProp],
	setValue func(ctx context.Context, dst *TObject, val TProp) error,
) *StructFieldVisitor[TObject, TProp] {
	return &StructFieldVisitor[TObject, TProp]{
		valueVisitor: propDec,
		setValue:     setValue,
	}
}

// Struct returns decoder to read YAML dictionary into structs.
func Struct[T any](props map[string]FieldVisitor[T]) *ObjectVisitor[T] {
	return &ObjectVisitor[T]{
		fields: props,
	}
}

// Map returns decoder to read YAML dictionaries into maps.
func Map[T any](dec ValueVisitor[T]) *MapVisitor[T] {
	return &MapVisitor[T]{
		handler: dec,
	}
}

// Misc:

// Reflect returns visitor that calls `yaml.Unmarshal` to unmarshal using reflection.
func Reflect[T any]() ValueVisitor[T] {
	return unmarshalVisitor[T]{}
}

// Any returns visitor which decodes unknown value using reflection.
func Any() ValueVisitor[any] {
	return unmarshalVisitor[any]{}
}

// Helpers:

// Optional writes default value if YAML node has empty value (ast.NullNode).
func Optional[T any](defaultVal T, dec ValueVisitor[T]) ValueVisitor[T] {
	return optionalVisitor[T]{
		defaultValue: defaultVal,
		valueVisitor: dec,
	}
}

// Duration returns visitor to decode a string into time.Duration using time.ParseDuration.
func Duration() ValueVisitor[time.Duration] {
	return durationVisitor{}
}

// VisitFunc returns visitor from a callback function.
func VisitFunc[T any](fn func(ctx context.Context, fi *TraverseOpts, node ast.Node) (T, parsetypes.Diagnostics)) ValueVisitor[T] {
	return visitorFunc[T]{
		fn: fn,
	}
}

// Validate wraps decoder with a value validation function.
func Validate[T any](dec ValueVisitor[T], validator func(ctx context.Context, t T) error) ValueVisitor[T] {
	return validatorVisitor[T]{
		decoder:      dec,
		validationFn: validator,
	}
}

// Transform wraps decoder with value transformer.
func Transform[TIn, TOut any](dec ValueVisitor[TIn], fn func(context.Context, ast.Node, TIn) (TOut, error)) ValueVisitor[TOut] {
	return transformVisitor[TIn, TOut]{
		dec:         dec,
		transformFn: fn,
	}
}
