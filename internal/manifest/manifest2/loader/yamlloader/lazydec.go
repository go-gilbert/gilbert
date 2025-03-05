package yamlloader

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

type lazyArrayVisitor struct{}

func (l lazyArrayVisitor) VisitItem(_ context.Context, opts *yamltree.TraverseOpts, node ast.Node) (*manifest2.LazyValue, parsetypes.Diagnostics) {
	switch t := node.(type) {
	case *ast.StringNode:
		// allow expressions
		v, _, diags := lazyFromStringNode(opts, t)
		if diags.HasError() {
			return nil, diags
		}

		if v.BindingSpec == nil {
			return nil, parsetypes.Diagnostics{
				yamltree.NewErrDiagnosticFromNode(opts.FileName, node, errors.New("expected expression in a string")),
			}
		}

		return &manifest2.LazyValue{
			Value:    v,
			Location: &v.BindingSpec.Location,
		}, nil
	case *ast.SequenceNode:
		v, _, diags := lazyFromSeqNode(opts, t)
		if diags.HasError() {
			return nil, diags
		}

		rng, offset := yamltree.GetNodeRange(node)
		return &manifest2.LazyValue{
			Value: v,
			Location: &manifest2.ReferenceLocation{
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
