package yamltree

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml/ast"
)

type listVisitor[T any] struct {
	listItemVisitor ValueVisitor[T]
}

func (v listVisitor[T]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) ([]T, parsetypes.Diagnostics) {
	if IsNullNode(node) {
		return nil, nil
	}

	var diags parsetypes.Diagnostics
	sq, ok := node.(*ast.SequenceNode)
	if !ok {
		diags = append(diags,
			newErrDiagnosticFromNode(
				opts.FileName, node,
				errors.New("node should be a list"),
			),
		)
		return nil, diags
	}

	out := make([]T, 0, len(sq.Values))
	for _, n := range sq.Values {
		if IsNullNode(n) {
			diags = append(diags,
				newErrDiagnosticFromNode(
					opts.FileName, n,
					errors.New("list item cannot be empty"),
				),
			)
			continue
		}

		r, d := v.listItemVisitor.VisitItem(ctx, opts, n)
		diags = append(diags, d...)

		if !d.HasError() {
			out = append(out, r)
		}
	}

	return out, diags
}

type ObjectVisitor[T any] struct {
	fields       []FieldVisitor[T]
	fieldsByName map[string]struct{}
	constructor  func(context.Context, *T) error
	validator    func(ctx context.Context, n ast.Node, dst *T) error
}

func (v *ObjectVisitor[T]) handleUnknownField(opts *TraverseOpts, key string, node *ast.MappingValueNode) *parsetypes.Diagnostic {
	if opts.UnknownFieldAction == UnknownFieldActionIgnore {
		return nil
	}

	diag := newErrDiagnosticFromMapping(opts, node, fmt.Errorf("unknown field %q", key))
	if opts.UnknownFieldAction == UnknownFieldActionWarn {
		diag.Severity = parsetypes.DiagnosticSeverityWarning
	}

	return diag
}

// Constructor initializes object using given func before mapping.
func (v *ObjectVisitor[T]) Constructor(fn func(context.Context, *T) error) *ObjectVisitor[T] {
	v.constructor = fn
	return v
}

// Validation adds object validation after all fields are mapped.
func (v *ObjectVisitor[T]) Validation(fn func(context.Context, ast.Node, *T) error) *ObjectVisitor[T] {
	v.validator = fn
	return v
}

func (v *ObjectVisitor[T]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) (T, parsetypes.Diagnostics) {
	nodes := make(map[string]*ast.MappingValueNode, len(v.fields))

	var (
		out   T
		diags parsetypes.Diagnostics
	)

	if v.constructor != nil {
		if err := v.constructor(ctx, &out); err != nil {
			diags = append(diags, newErrDiagnosticFromNode(opts.FileName, node, err))
		}
	}

	mn, ok := node.(*ast.MappingNode)
	if !ok {
		diags = append(diags,
			newErrDiagnosticFromNode(
				opts.FileName, node,
				fmt.Errorf("node should be an object, got %s", node.Type()),
			),
		)

		return out, diags
	}

	// Collect field members & validate key names
	for _, n := range mn.Values {
		kn, ok := n.Key.(*ast.StringNode)
		if !ok {
			diags = append(diags,
				newErrDiagnosticFromMapping(opts, n, errors.New("key should be a string")),
			)

			continue
		}

		if _, ok := v.fieldsByName[kn.Value]; !ok {
			if diag := v.handleUnknownField(opts, kn.Value, n); diag != nil {
				diags = append(diags, diag)
			}
			continue
		}

		if _, ok := nodes[kn.Value]; ok {
			diags = append(diags,
				newErrDiagnosticFromMapping(opts, n, fmt.Errorf("duplicate field %q", kn.Value)),
			)
			continue
		}

		nodes[kn.Value] = n
	}

	// Map fields respecting schema order
	for _, dec := range v.fields {
		key := dec.Name()
		n, ok := nodes[key]
		if !ok {
			if dec.IsRequired() {
				diags = append(diags,
					newErrDiagnosticFromNode(
						opts.FileName, node,
						fmt.Errorf("field %q is required", key),
					),
				)
			}
			continue
		}

		fieldDiags := dec.VisitField(ctx, opts, n, &out)
		diags = append(diags, fieldDiags...)
		if fieldDiags.HasError() {
			continue
		}

		if err := dec.Validate(ctx, &out); err != nil {
			diags = append(diags,
				newErrDiagnosticFromNode(opts.FileName, n, err),
			)
		}
	}

	if !diags.HasError() && v.validator != nil {
		if err := v.validator(ctx, node, &out); err != nil {
			diags = append(diags, newErrDiagnosticFromNode(opts.FileName, node, err))
		}
	}

	return out, diags
}

type MapVisitor[T any] struct {
	handler              ValueVisitor[T]
	keyTransformer       func(context.Context, string) (string, error)
	duplicateItemHandler func(context.Context, string, T) error
	docHandler           func(context.Context, FieldInfo, T) T
}

// OnDuplicateKey sets a function to format duplicate key errors.
func (v *MapVisitor[T]) OnDuplicateKey(fn func(context.Context, string, T) error) *MapVisitor[T] {
	v.duplicateItemHandler = fn
	return v
}

// TransformKey set a function to transform dictionary keys during mapping.
func (v *MapVisitor[T]) TransformKey(fn func(context.Context, string) (string, error)) *MapVisitor[T] {
	v.keyTransformer = fn
	return v
}

// CollectDoc sets handler to collect field documentation.
func (v *MapVisitor[T]) CollectDoc(fn func(context.Context, FieldInfo, T) T) *MapVisitor[T] {
	v.docHandler = fn
	return v
}

func (v *MapVisitor[T]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) (map[string]T, parsetypes.Diagnostics) {
	if IsNullNode(node) {
		return nil, nil
	}

	mn, diags := intoDictNode(opts, node)
	if len(diags) != 0 {
		return nil, diags
	}

	m := make(map[string]T, len(mn.Values))
	diags = v.readNode(ctx, opts, mn, m)
	return m, diags
}

func (v *MapVisitor[T]) VisitNodeInto(ctx context.Context, opts *TraverseOpts, node ast.Node, dst map[string]T) parsetypes.Diagnostics {
	if node.Type() == ast.NullType {
		return nil
	}

	mn, diags := intoDictNode(opts, node)
	if len(diags) != 0 {
		return diags
	}

	return v.readNode(ctx, opts, mn, dst)
}

func (v *MapVisitor[T]) readNode(ctx context.Context, opts *TraverseOpts, node *ast.MappingNode, dst map[string]T) parsetypes.Diagnostics {
	var diags parsetypes.Diagnostics
	for _, n := range node.Values {
		diags = append(diags, v.visitChild(ctx, opts, n, dst)...)
	}

	return diags
}

func (v *MapVisitor[T]) formatDuplicateError(ctx context.Context, key string, prev T) error {
	if v.duplicateItemHandler != nil {
		return v.duplicateItemHandler(ctx, key, prev)
	}

	return fmt.Errorf("duplicate key %q", key)
}

func (v *MapVisitor[T]) formatKey(ctx context.Context, k string) (string, error) {
	if v.keyTransformer == nil {
		return k, nil
	}

	return v.keyTransformer(ctx, k)
}

func (v *MapVisitor[T]) visitChild(ctx context.Context, opts *TraverseOpts, n *ast.MappingValueNode, dst map[string]T) parsetypes.Diagnostics {
	var diags parsetypes.Diagnostics
	kn, ok := n.Key.(*ast.StringNode)
	if !ok {
		diags = append(diags,
			newErrDiagnosticFromNode(
				opts.FileName, kn,
				errors.New("key should be a string"),
			),
		)
		return diags
	}

	key, err := v.formatKey(ctx, kn.Value)
	if err != nil {
		diags = append(diags, newErrDiagnosticFromNode(opts.FileName, kn, err))
		return diags
	}

	if prev, ok := dst[key]; ok {
		diags = append(diags,
			newErrDiagnosticFromNode(
				opts.FileName, kn, v.formatDuplicateError(ctx, key, prev)),
		)
		return diags
	}

	item, d := v.handler.VisitItem(ctx, opts, n.Value)
	diags = append(diags, d...)
	if !d.HasError() {
		if v.docHandler != nil {
			fi := FieldInfo{
				Key: key,
				Doc: collectFieldDoc(n),
			}

			item = v.docHandler(ctx, fi, item)
		}

		dst[key] = item
	}

	return diags
}
