package yamltree

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml/ast"
)

type visitorFunc[T any] struct {
	fn func(ctx context.Context, opts *TraverseOpts, node ast.Node) (T, parsetypes.Diagnostics)
}

func (v visitorFunc[T]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) (T, parsetypes.Diagnostics) {
	return v.fn(ctx, opts, node)
}

type validatorVisitor[T any] struct {
	decoder      ValueVisitor[T]
	validationFn func(ctx context.Context, t T) error
}

func (v validatorVisitor[T]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) (T, parsetypes.Diagnostics) {
	val, diags := v.decoder.VisitItem(ctx, opts, node)
	if diags.HasError() {
		return val, diags
	}

	if err := v.validationFn(ctx, val); err != nil {
		diags = append(diags, newErrDiagnosticFromNode(opts.FileName, node, err))
	}

	return val, diags
}

type durationVisitor struct{}

func (_ durationVisitor) VisitItem(_ context.Context, fi *TraverseOpts, node ast.Node) (time.Duration, parsetypes.Diagnostics) {
	if IsNullNode(node) {
		return 0, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node,
				errors.New("missing duration value"),
			),
		}
	}

	var n *ast.StringNode
	switch v := node.(type) {
	case *ast.StringNode:
		n = v
	case *ast.LiteralNode:
		n = v.Value
	default:
		return 0, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node,
				fmt.Errorf("expected duration string, got %s", v.Type()),
			),
		}
	}

	val := strings.TrimSpace(n.Value)
	dur, err := time.ParseDuration(val)
	if err != nil {
		return 0, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, n,
				fmt.Errorf("invalid duration string format: %w", err),
			),
		}
	}

	return dur, nil
}

type optionalVisitor[T any] struct {
	valueVisitor ValueVisitor[T]
	defaultValue T
}

func (v optionalVisitor[T]) VisitItem(ctx context.Context, fi *TraverseOpts, node ast.Node) (T, parsetypes.Diagnostics) {
	if node == nil || node.Type() == ast.NullType {
		return v.defaultValue, nil
	}

	return v.valueVisitor.VisitItem(ctx, fi, node)
}

type transformVisitor[TIn, TOut any] struct {
	dec         ValueVisitor[TIn]
	transformFn func(context.Context, ast.Node, TIn) (TOut, error)
}

func (tv transformVisitor[TIn, TOut]) VisitItem(ctx context.Context, fi *TraverseOpts, node ast.Node) (out TOut, diags parsetypes.Diagnostics) {
	val, diags := tv.dec.VisitItem(ctx, fi, node)
	if diags.HasError() {
		return out, diags
	}

	next, err := tv.transformFn(ctx, node, val)
	if err != nil {
		diags = append(diags, newErrDiagnosticFromNode(fi.FileName, node, err))
	}

	return next, diags
}
