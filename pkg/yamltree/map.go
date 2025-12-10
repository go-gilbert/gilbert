package yamltree

import (
	"context"
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type MapKeyRef struct {
	Opts    *TraverseOpts
	KeyNode *ast.StringNode
}

type MapCollector[TResult, TElem any] struct {
	MakeNew     func(elemCount int) TResult
	FormatKey   func(ctx context.Context, kref MapKeyRef, dst TResult) (string, *parsetypes.Diagnostic)
	AppendValue func(ctx context.Context, dst TResult, key string, v TElem) (TResult, error)
}

type MapIterator[TResult, TElem any] struct {
	valDecoder ValueVisitor[TElem]
	collector  MapCollector[TResult, TElem]
	docHandler func(context.Context, FieldInfo, TElem) TElem
}

// CollectDoc sets handler to collect field documentation.
func (v *MapIterator[TResult, TElem]) CollectDoc(fn func(context.Context, FieldInfo, TElem) TElem) *MapIterator[TResult, TElem] {
	v.docHandler = fn
	return v
}

func (v *MapIterator[TResult, TElem]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) (TResult, parsetypes.Diagnostics) {
	var r TResult
	if IsNullNode(node) {
		return r, nil
	}

	mn, diags := intoDictNode(opts, node)
	if len(diags) != 0 {
		return r, diags
	}

	dst := v.collector.MakeNew(len(mn.Values))
	for _, n := range mn.Values {
		r, d := v.visitChild(ctx, opts, n, dst)
		diags = append(diags, d...)
		dst = r
	}

	return dst, diags
}

func (v *MapIterator[TResult, TElem]) visitChild(ctx context.Context, opts *TraverseOpts, n *ast.MappingValueNode, acc TResult) (TResult, parsetypes.Diagnostics) {
	var diags parsetypes.Diagnostics
	kn, ok := n.Key.(*ast.StringNode)
	if !ok {
		diags = append(diags,
			NewErrDiagnosticFromNode(
				opts.FileName, kn,
				errors.New("key should be a string"),
			),
		)
		return acc, diags
	}

	key := kn.Value
	if v.collector.FormatKey != nil {
		kref := MapKeyRef{
			Opts:    opts,
			KeyNode: kn,
		}

		var kDiag *parsetypes.Diagnostic
		key, kDiag = v.collector.FormatKey(ctx, kref, acc)
		if kDiag != nil {
			diags = append(diags, kDiag)
			return acc, diags
		}
	}

	item, d := v.valDecoder.VisitItem(ctx, opts, n.Value)
	diags = append(diags, d...)
	if d.HasError() {
		return acc, diags
	}

	if v.docHandler != nil {
		fi := FieldInfo{
			Key: key,
			Doc: collectFieldDoc(n),
		}

		item = v.docHandler(ctx, fi, item)
	}

	result, err := v.collector.AppendValue(ctx, acc, key, item)
	if err != nil {
		diag, ok := parsetypes.DiagnosticFromError(err)
		if !ok {
			diag = NewErrDiagnosticFromNode(opts.FileName, kn, err)
		}
		diags = append(diags, diag)
		return acc, diags
	}

	return result, diags
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

// func (v *MapVisitor[T]) VisitNodeInto(ctx context.Context, opts *TraverseOpts, node ast.Node, dst map[string]T) parsetypes.Diagnostics {
// 	if node.Type() == ast.NullType {
// 		return nil
// 	}
//
// 	mn, diags := intoDictNode(opts, node)
// 	if len(diags) != 0 {
// 		return diags
// 	}
//
// 	return v.readNode(ctx, opts, mn, dst)
// }

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
			NewErrDiagnosticFromNode(
				opts.FileName, kn,
				errors.New("key should be a string"),
			),
		)
		return diags
	}

	key, err := v.formatKey(ctx, kn.Value)
	if err != nil {
		diags = append(diags, NewErrDiagnosticFromNode(opts.FileName, kn, err))
		return diags
	}

	if prev, ok := dst[key]; ok {
		diags = append(diags,
			NewErrDiagnosticFromNode(
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
