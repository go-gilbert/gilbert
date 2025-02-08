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

func (v listVisitor[T]) VisitItem(ctx context.Context, fi FileInfo, node ast.Node) ([]T, parsetypes.Diagnostics) {
	if IsNullNode(node) {
		return nil, nil
	}

	var diags parsetypes.Diagnostics
	sq, ok := node.(*ast.SequenceNode)
	if !ok {
		diags = append(diags,
			newErrDiagnosticFromNode(
				fi.FileName, node,
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
					fi.FileName, n,
					errors.New("list item cannot be empty"),
				),
			)
		}

		r, d := v.listItemVisitor.VisitItem(ctx, fi, n)
		diags = append(diags, d...)

		if !d.HasError() {
			out = append(out, r)
		}
	}

	return out, diags
}

type FieldVisitor[T any] interface {
	IsRequired() bool
	VisitItem(ctx context.Context, fi FileInfo, node ast.Node, dst *T) parsetypes.Diagnostics
}

type fieldVisitor[TObject, TProp any] struct {
	required     bool
	valueVisitor ValueVisitor[TProp]
	setValue     func(ctx context.Context, dst *TObject, val TProp) error
}

func (fv *fieldVisitor[TObject, TProp]) IsRequired() bool {
	return fv.required
}

func (fv *fieldVisitor[TObject, TProp]) VisitItem(ctx context.Context, fi FileInfo, node ast.Node, dst *TObject) parsetypes.Diagnostics {
	if fv.setValue == nil {
		panic("propertyVisitor: missing value setter")
	}

	v, diags := fv.valueVisitor.VisitItem(ctx, fi, node)
	if diags.HasError() {
		return diags
	}

	if err := fv.setValue(ctx, dst, v); err != nil {
		diags = append(diags, newErrDiagnosticFromNode(fi.FileName, node, err))
	}

	return diags
}

type objectVisitor[T any] struct {
	strict      bool
	fields      map[string]FieldVisitor[T]
	validatorFn func(ctx context.Context, v T) error
}

func (v *objectVisitor[T]) VisitItem(ctx context.Context, fi FileInfo, node ast.Node) (T, parsetypes.Diagnostics) {
	visitedFields := make(map[string]struct{}, len(v.fields))

	var (
		out   T
		diags parsetypes.Diagnostics
	)

	mn, ok := node.(*ast.MappingNode)
	if !ok {
		diags = append(diags,
			newErrDiagnosticFromNode(
				fi.FileName, node,
				errors.New("node should be an object"),
			),
		)

		return out, diags
	}

	for _, n := range mn.Values {
		kn, ok := n.Key.(*ast.StringNode)
		if !ok {
			diags = append(diags,
				newErrDiagnosticFromMapping(fi, n, errors.New("key should be a string")),
			)

			continue
		}

		fv, ok := v.fields[kn.Value]
		if !ok {
			if v.strict {
				diags = append(diags,
					newErrDiagnosticFromMapping(fi, n, fmt.Errorf("unknown field %q", kn.Value)),
				)
			}
			continue
		}

		if _, ok := visitedFields[kn.Value]; ok {
			diags = append(diags,
				newErrDiagnosticFromMapping(fi, n, fmt.Errorf("duplicate field %q", kn.Value)),
			)
			continue
		}

		visitedFields[kn.Value] = struct{}{}
		diags = append(diags, fv.VisitItem(ctx, fi, n.Value, &out)...)
	}

	for key, f := range v.fields {
		if !f.IsRequired() {
			continue
		}

		if _, ok := visitedFields[key]; !ok {
			diags = append(diags,
				newErrDiagnosticFromNode(
					fi.FileName, node,
					fmt.Errorf("field %q is required", key),
				),
			)
		}
	}

	if v.validatorFn != nil && !diags.HasError() {
		if err := v.validatorFn(ctx, out); err != nil {
			diags = append(diags, newErrDiagnosticFromNode(fi.FileName, node, err))
		}
	}

	return out, diags
}

type dictVisitor[T any] struct {
	handler              ValueVisitor[T]
	keyTransformer       func(context.Context, string) (string, error)
	duplicateItemHandler func(context.Context, string, T) error
}

func (v *dictVisitor[T]) VisitItem(ctx context.Context, fi FileInfo, node ast.Node) (map[string]T, parsetypes.Diagnostics) {
	if IsNullNode(node) {
		return nil, nil
	}

	mn, diags := intoDictNode(fi, node)
	if len(diags) != 0 {
		return nil, diags
	}

	m := make(map[string]T, len(mn.Values))
	diags = v.readNode(ctx, fi, mn, m)
	return m, diags
}

func (v *dictVisitor[T]) VisitNodeInto(ctx context.Context, fi FileInfo, node ast.Node, dst map[string]T) parsetypes.Diagnostics {
	if node.Type() == ast.NullType {
		return nil
	}

	mn, diags := intoDictNode(fi, node)
	if len(diags) != 0 {
		return diags
	}

	return v.readNode(ctx, fi, mn, dst)
}

func (v *dictVisitor[T]) readNode(ctx context.Context, fi FileInfo, node *ast.MappingNode, dst map[string]T) parsetypes.Diagnostics {
	var diags parsetypes.Diagnostics
	for _, n := range node.Values {
		diags = append(diags, v.visitChild(ctx, fi, n, dst)...)
	}

	return diags
}

func (v *dictVisitor[T]) formatDuplicateError(ctx context.Context, key string, prev T) error {
	if v.duplicateItemHandler != nil {
		return v.duplicateItemHandler(ctx, key, prev)
	}

	return fmt.Errorf("duplicate key %q", key)
}

func (v *dictVisitor[T]) formatKey(ctx context.Context, k string) (string, error) {
	if v.keyTransformer == nil {
		return k, nil
	}

	return v.keyTransformer(ctx, k)
}

func (v *dictVisitor[T]) visitChild(ctx context.Context, fi FileInfo, n *ast.MappingValueNode, dst map[string]T) parsetypes.Diagnostics {
	var diags parsetypes.Diagnostics
	kn, ok := n.Key.(*ast.StringNode)
	if !ok {
		diags = append(diags,
			newErrDiagnosticFromNode(
				fi.FileName, kn,
				errors.New("key should be a string"),
			),
		)
		return diags
	}

	key, err := v.formatKey(ctx, kn.Value)
	if err != nil {
		diags = append(diags, newErrDiagnosticFromNode(fi.FileName, kn, err))
		return diags
	}

	if prev, ok := dst[key]; ok {
		diags = append(diags,
			newErrDiagnosticFromNode(
				fi.FileName, kn, v.formatDuplicateError(ctx, key, prev)),
		)
		return diags
	}

	item, d := v.handler.VisitItem(ctx, fi, n)
	diags = append(diags, d...)
	if !d.HasError() {
		dst[key] = item
	}

	return diags
}
