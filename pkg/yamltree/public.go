// Package yamltree provides tools for unmarshaling and validation of YAML AST trees without reflection.
package yamltree

import (
	"context"
	"time"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml/ast"
	"golang.org/x/exp/constraints"
)

type FileInfo struct {
	FileName string
}

// ValueVisitor is interface to implement decoding YAML AST nodes to values.
type ValueVisitor[T any] interface {
	VisitItem(ctx context.Context, fi FileInfo, node ast.Node) (T, parsetypes.Diagnostics)
}

// Primitives:

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

// Arrays:

// List returns decoder for arrays.
func List[T any](itemDec ValueVisitor[T]) ValueVisitor[[]T] {
	return listVisitor[T]{
		listItemVisitor: itemDec,
	}
}

// Structs:

// Field returns decoder for object property.
//
// Meant to be passed inside Struct().
func Field[TObject, TProp any](
	isRequired bool,
	propDec ValueVisitor[TProp],
	setValue func(ctx context.Context, dst *TObject, val TProp) error,
) FieldVisitor[TObject] {
	return &fieldVisitor[TObject, TProp]{
		valueVisitor: propDec,
		required:     isRequired,
		setValue:     setValue,
	}
}

// WithStructStrictFields option enables error report on unknown object property.
func WithStructStrictFields[T any]() func(opts *objectVisitor[T]) {
	return func(o *objectVisitor[T]) {
		o.strict = true
	}
}

// WithStructValidation validates object produced by Struct().
func WithStructValidation[T any](fn func(ctx context.Context, v T) error) func(opts *objectVisitor[T]) {
	return func(o *objectVisitor[T]) {
		o.validatorFn = fn
	}
}

// Struct returns decoder to read YAML dictionary into structs.
func Struct[T any](props map[string]FieldVisitor[T], opts ...func(*objectVisitor[T])) ValueVisitor[T] {
	v := &objectVisitor[T]{
		fields: props,
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Maps:

// WithMapKeyTransformer transforms and/or validates map keys before insertion.
func WithMapKeyTransformer[T any](fn func(context.Context, string) (string, error)) func(*dictVisitor[T]) {
	return func(o *dictVisitor[T]) {
		o.keyTransformer = fn
	}
}

// WithMapDuplicateKeyHandler allows to override customise duplicate map key error.
func WithMapDuplicateKeyHandler[T any](fn func(context.Context, string, T) error) func(*dictVisitor[T]) {
	return func(o *dictVisitor[T]) {
		o.duplicateItemHandler = fn
	}
}

// Map returns decoder to read YAML dictionaries into maps.
func Map[T any](dec ValueVisitor[T], opts ...func(*dictVisitor[T])) ValueVisitor[map[string]T] {
	dv := &dictVisitor[T]{
		handler: dec,
	}

	for _, opt := range opts {
		opt(dv)
	}

	return dv
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
func VisitFunc[T any](fn func(ctx context.Context, fi FileInfo, node ast.Node) (T, parsetypes.Diagnostics)) ValueVisitor[T] {
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
func Transform[T any](dec ValueVisitor[T], fn func(context.Context, T) (T, error)) ValueVisitor[T] {
	return transformVisitor[T]{
		dec:         dec,
		transformFn: fn,
	}
}
