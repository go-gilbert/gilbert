package yamlloader

import (
	"context"
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"

	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/go-gilbert/gilbert/pkg/yamltree"
)

type scalarOrExprVisitor struct{}

// ScalarOrExpression returns a visitor for scalar YAML node or an expression.
func ScalarOrExpression() yamltree.ValueVisitor[*manifest.LazyValue] {
	return scalarOrExprVisitor{}
}

func (vis scalarOrExprVisitor) VisitItem(ctx context.Context, opts *yamltree.TraverseOpts, node ast.Node) (*manifest.LazyValue, parsetypes.Diagnostics) {
	switch t := node.(type) {
	case *ast.StringNode:
		// allow expressions
		v, _, diags := lazyFromStringNode(opts, t, nil)
		if diags.HasError() {
			return nil, diags
		}

		return &manifest.LazyValue{
			Value:    v.Optimize(),
			Location: &v.BindingSpec.Location,
		}, nil
	case ast.ScalarNode:
		v := t.GetValue()
		rng, offset := yamltree.GetNodeRange(node)
		return &manifest.LazyValue{
			Value: manifest.AnySpec{
				LiteralSpec: &manifest.LiteralSpec{
					Value: v,
				},
			},
			Location: &manifest.ReferenceLocation{
				FileName: opts.FileName,
				Range:    rng,
				Offset:   offset,
			},
		}, nil
	}

	return nil, parsetypes.Diagnostics{
		yamltree.NewErrDiagnosticFromNode(opts.FileName, node, errors.New("expected expression or scalar value")),
	}
}

type lazyArrayVisitor struct{}

func (l lazyArrayVisitor) VisitItem(_ context.Context, opts *yamltree.TraverseOpts, node ast.Node) (*manifest.LazyValue, parsetypes.Diagnostics) {
	switch t := node.(type) {
	case *ast.StringNode:
		// allow expressions
		v, _, diags := lazyFromStringNode(opts, t, nil)
		if diags.HasError() {
			return nil, diags
		}

		if v.BindingSpec == nil {
			return nil, parsetypes.Diagnostics{
				yamltree.NewErrDiagnosticFromNode(opts.FileName, node, errors.New("expected expression in a string")),
			}
		}

		return &manifest.LazyValue{
			Value:    v.Optimize(),
			Location: &v.BindingSpec.Location,
		}, nil
	case *ast.SequenceNode:
		v, _, diags := lazyFromSeqNode(opts, t)
		if diags.HasError() {
			return nil, diags
		}

		rng, offset := yamltree.GetNodeRange(node)
		return &manifest.LazyValue{
			Value: v.Optimize(),
			Location: &manifest.ReferenceLocation{
				FileName: opts.FileName,
				Range:    rng,
				Offset:   offset,
			},
		}, nil
	default:
		return nil, parsetypes.Diagnostics{
			yamltree.NewErrDiagnosticFromNode(opts.FileName, node, fmt.Errorf("value should be array or expression, got %s", t)),
		}
	}
}
