package yamltree

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

type FieldInfo struct {
	Key string
	Doc []string
}

type FieldVisitor[T any] interface {
	// Name returns YAML key name.
	Name() string

	// IsRequired returns whether field is required.
	IsRequired() bool

	// Validate validates field value.
	Validate(ctx context.Context, dst *T) error

	// VisitField decodes object field.
	VisitField(ctx context.Context, opts *TraverseOpts, node *ast.MappingValueNode, dst *T) parsetypes.Diagnostics
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

func (f *FuncFieldVisitor[TObject, TProp]) VisitField(ctx context.Context, opts *TraverseOpts, node *ast.MappingValueNode, dst *TObject) parsetypes.Diagnostics {
	v, err := f.visitorFunc(ctx, node.Value, dst)
	if err != nil {
		return parsetypes.Diagnostics{
			NewErrDiagnosticFromNode(opts.FileName, node, err),
		}
	}

	result, diags := v.VisitItem(ctx, opts, node.Value)
	if diags.HasError() {
		return diags
	}

	if err := f.setValue(ctx, dst, result); err != nil {
		diags = append(diags, NewErrDiagnosticFromNode(opts.FileName, node.Value, err))
	}

	return diags
}

// Required marks field a required.
func (f *FuncFieldVisitor[TObject, TProp]) Required() *FuncFieldVisitor[TObject, TProp] {
	f.required = true
	return f
}

type StructFieldVisitor[TObject, TProp any] struct {
	name          string
	required      bool
	valueVisitor  ValueVisitor[TProp]
	validator     func(ctx context.Context, dst *TObject) error
	setValue      func(ctx context.Context, dst *TObject, val TProp) error
	docReaderFunc func(ctx context.Context, fi FieldInfo, dst *TObject)
	nodeFunc      func(ctx context.Context, n *ast.MappingValueNode, dst *TObject)
	ctxFunc       func(ctx context.Context) context.Context
}

// Required marks struct field as required.
func (fv *StructFieldVisitor[TObject, TProp]) Required() *StructFieldVisitor[TObject, TProp] {
	fv.required = true
	return fv
}

// CollectDoc sets handler to collect field documentation.
func (fv *StructFieldVisitor[TObject, TProp]) CollectDoc(fn func(ctx context.Context, fi FieldInfo, dst *TObject)) {
	fv.docReaderFunc = fn
}

// Validation adds validation func to be called to validate field value.
func (fv *StructFieldVisitor[TObject, TProp]) Validation(fn func(context.Context, *TObject) error) *StructFieldVisitor[TObject, TProp] {
	fv.validator = fn
	return fv
}

// CheckNode adds a hook to processs raw value mapping AST node.
func (fv *StructFieldVisitor[TObject, TProp]) CheckNode(fn func(context.Context, *ast.MappingValueNode, *TObject)) *StructFieldVisitor[TObject, TProp] {
	fv.nodeFunc = fn
	return fv
}

// WithContext adds a function to wrap context with custom value during decoding.
func (fv *StructFieldVisitor[TObject, TProp]) WithContext(fn func(ctx context.Context) context.Context) *StructFieldVisitor[TObject, TProp] {
	fv.ctxFunc = fn
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

// VisitField implements FieldVisitor interface.
func (fv *StructFieldVisitor[TObject, TProp]) VisitField(ctx context.Context, opts *TraverseOpts, node *ast.MappingValueNode, dst *TObject) parsetypes.Diagnostics {
	if fv.setValue == nil {
		panic(fmt.Sprintf("StructFieldVisitor: missing value setter for field %q", fv.name))
	}

	if fv.valueVisitor == nil {
		panic(fmt.Sprintf("StructFieldVisitor: missing value decoder for field %q", fv.name))
	}

	if fv.ctxFunc != nil {
		ctx = fv.ctxFunc(ctx)
	}

	v, diags := fv.valueVisitor.VisitItem(ctx, opts, node.Value)
	if diags.HasError() {
		return diags
	}

	if fv.docReaderFunc != nil {
		fi := FieldInfo{
			Key: node.Key.(*ast.StringNode).Value,
			Doc: collectFieldDoc(node),
		}

		fv.docReaderFunc(ctx, fi, dst)
	}

	if fv.nodeFunc != nil {
		fv.nodeFunc(ctx, node, dst)
	}

	if err := fv.setValue(ctx, dst, v); err != nil {
		diags = append(diags, NewErrDiagnosticFromNode(opts.FileName, node.Value, err))
	}

	return diags
}

// collectFieldDoc collects comments defined above a node.
//
// Although ast.Node supposed to have a comment, the go-yaml lib actually doesn't populate that info.
func collectFieldDoc(n *ast.MappingValueNode) []string {
	keyTok := n.Key.GetToken()
	prevLine := keyTok.Position.Line

	var doc []string
	prevTok := keyTok.Prev
	for {
		if prevTok == nil || prevTok.Type != token.CommentType {
			// stop at non-comment token
			break
		}

		curLine := prevTok.Position.Line
		if prevLine-curLine > 1 {
			// Comment block finish
			break
		}

		doc = append(doc, strings.TrimSpace(prevTok.Value))
		prevLine = curLine
		prevTok = prevTok.Prev
	}

	if len(doc) > 1 {
		slices.Reverse(doc)
	}

	return doc
}
