package yamltree

import (
	"context"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml/ast"
)

type FieldVisitor[T any] interface {
	// Name returns YAML key name.
	Name() string

	// IsRequired returns whether field is required.
	IsRequired() bool

	// Validate validates field value.
	Validate(ctx context.Context, dst *T) error

	// VisitItem decodes field value.
	VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node, dst *T) parsetypes.Diagnostics
}

type FuncFieldVisitor[TObject, TProp any] struct {
	name        string
	required    bool
	visitorFunc func(ctx context.Context, node ast.Node, dst *TObject) (ValueVisitor[TProp], error)
	setValue    func(ctx context.Context, dst *TObject, val TProp) error
}

func (f *FuncFieldVisitor[TObject, TProp]) Name() string {
	return f.name
}

func (f *FuncFieldVisitor[TObject, TProp]) IsRequired() bool {
	return f.required
}

func (f *FuncFieldVisitor[TObject, TProp]) Validate(ctx context.Context, dst *TObject) error {
	// TODO
	return nil
}

func (f *FuncFieldVisitor[TObject, TProp]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node, dst *TObject) parsetypes.Diagnostics {
	v, err := f.visitorFunc(ctx, node, dst)
	if err != nil {
		return parsetypes.Diagnostics{
			newErrDiagnosticFromNode(opts.FileName, node, err),
		}
	}

	result, diags := v.VisitItem(ctx, opts, node)
	if diags.HasError() {
		return diags
	}

	if err := f.setValue(ctx, dst, result); err != nil {
		diags = append(diags, newErrDiagnosticFromNode(opts.FileName, node, err))
	}

	return diags
}

// Required marks field a required.
func (f *FuncFieldVisitor[TObject, TProp]) Required() *FuncFieldVisitor[TObject, TProp] {
	f.required = true
	return f
}

type StructFieldVisitor[TObject, TProp any] struct {
	name         string
	required     bool
	valueVisitor ValueVisitor[TProp]
	validator    func(ctx context.Context, dst *TObject) error
	setValue     func(ctx context.Context, dst *TObject, val TProp) error
}

// Required marks struct field as required.
func (fv *StructFieldVisitor[TObject, TProp]) Required() *StructFieldVisitor[TObject, TProp] {
	fv.required = true
	return fv
}

// Validation adds validation func to be called to validate field value.
func (fv *StructFieldVisitor[TObject, TProp]) Validation(fn func(context.Context, *TObject) error) *StructFieldVisitor[TObject, TProp] {
	fv.validator = fn
	return fv
}

// Name implements FieldVisitor interface.
func (fv *StructFieldVisitor[TObject, TProp]) Name() string {
	return fv.name
}

// IsRequired implements FieldVisitor interface.
func (fv *StructFieldVisitor[TObject, TProp]) IsRequired() bool {
	return fv.required
}

// Validate implements FieldVisitor interface.
func (fv *StructFieldVisitor[TObject, TProp]) Validate(ctx context.Context, dst *TObject) error {
	if fv.validator != nil {
		return fv.validator(ctx, dst)
	}

	return nil
}

// VisitItem implements FieldVisitor interface.
func (fv *StructFieldVisitor[TObject, TProp]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node, dst *TObject) parsetypes.Diagnostics {
	if fv.setValue == nil {
		panic("propertyVisitor: missing value setter")
	}

	v, diags := fv.valueVisitor.VisitItem(ctx, opts, node)
	if diags.HasError() {
		return diags
	}

	if err := fv.setValue(ctx, dst, v); err != nil {
		diags = append(diags, newErrDiagnosticFromNode(opts.FileName, node, err))
	}

	return diags
}
