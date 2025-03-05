package yamlloader

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-gilbert/gilbert/internal/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

type endPosition struct {
	offset int
	cursor parsetypes.Position
}

func (endPos endPosition) referenceLocation(opts *yamltree.TraverseOpts, startTok *token.Token) *manifest2.ReferenceLocation {
	startPos := startTok.Position
	return &manifest2.ReferenceLocation{
		FileName: opts.FileName,
		Range: parsetypes.NewRange(
			parsetypes.NewPosition(startPos.Line, startPos.Column),
			endPos.cursor,
		),
		Offset: parsetypes.OffsetRange{
			Start: startPos.Offset,
			End:   endPos.offset,
		},
	}
}

var _ yamltree.ValueVisitor[*manifest2.LazyValue] = (*lazyValueVisitor)(nil)

type lazyValueVisitor struct{}

func (_ lazyValueVisitor) VisitItem(_ context.Context, opts *yamltree.TraverseOpts, node ast.Node) (*manifest2.LazyValue, parsetypes.Diagnostics) {
	v, endPos, diags := lazyFromNode(opts, node)
	if !diags.HasError() {
		v = v.Optimize()
	}

	return &manifest2.LazyValue{
		Location: endPos.referenceLocation(opts, node.GetToken()),
		Value:    v,
	}, diags
}

func lazyFromNode(opts *yamltree.TraverseOpts, node ast.Node) (rawNode manifest2.AnySpec, endPos endPosition, diags parsetypes.Diagnostics) {
	endPos = endPositionFromToken(node.GetToken())
	if yamltree.IsNullNode(node) {
		diags = append(diags,
			newErrDiagnosticFromNode(
				opts.FileName, node,
				errors.New("list item cannot be empty"),
			),
		)
		return rawNode, endPos, diags
	}

	switch n := node.(type) {
	case *ast.SequenceNode:
		return lazyFromSeqNode(opts, n)
	case *ast.MappingNode:
		return lazyFromDictNode(opts, n)
	case *ast.StringNode:
		return lazyFromStringNode(opts, n)
	case *ast.LiteralNode:
		return lazyFromLiteralNode(opts, n)
	case *ast.IntegerNode:
		rawNode, endPos = lazyFromVal(n.Token, n.Value)
	case *ast.FloatNode:
		rawNode, endPos = lazyFromVal(n.Token, n.Value)
	case *ast.BoolNode:
		rawNode, endPos = lazyFromVal(n.Token, n.Value)
	default:
		diags = append(diags,
			newErrDiagnosticFromNode(opts.FileName, node, fmt.Errorf("unsupported node: %s", node.Type())),
		)
	}

	return rawNode, endPos, diags
}

func lazyFromLiteralNode(opts *yamltree.TraverseOpts, node *ast.LiteralNode) (manifest2.AnySpec, endPosition, parsetypes.Diagnostics) {
	return lazyFromStringNode(opts, node.Value)
	//s, diags := lazyFromStringNode(opts, node.Value)
	//if diags.HasError() {
	//	return s, diags
	//}

	//if s.IsLiteral() {
	//	return s, diags
	//}
}

func lazyFromStringNode(opts *yamltree.TraverseOpts, node *ast.StringNode) (manifest2.AnySpec, endPosition, parsetypes.Diagnostics) {
	endPos := endPositionFromToken(node.Token)
	if node.Value == "" {
		return manifest2.AnySpec{
			LiteralSpec: &manifest2.LiteralSpec{
				Value: "",
			},
		}, endPos, nil
	}

	// TODO: unwrap
	e, err := expr.Parse(node.Value)
	if err != nil {
		return manifest2.AnySpec{}, endPos, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(opts.FileName, node, err),
		}
	}

	if !e.Evaluable() {
		return manifest2.AnySpec{
			LiteralSpec: &manifest2.LiteralSpec{
				Value: node.Value,
			},
		}, endPos, nil
	}

	// FIXME: set correct bounds for multiline strings from *ast.LiteralNode.
	pos := parsetypes.NewPosition(node.Token.Position.Line, node.Token.Position.Column)
	return manifest2.AnySpec{
		BindingSpec: &manifest2.BindingSpec{
			Expr: e,
			Location: manifest2.ReferenceLocation{
				FileName: opts.FileName,
				Range:    parsetypes.NewRange(pos, pos.Add(0, len(node.Value)-1)),
				Offset:   parsetypes.NewOffsetRange(node.Token.Position.Offset, len(node.Token.Origin)-1),
			},
		},
	}, endPos, nil
}

func lazyFromVal(tok *token.Token, val any) (manifest2.AnySpec, endPosition) {
	endPos := endPositionFromToken(tok)

	return manifest2.AnySpec{
		LiteralSpec: &manifest2.LiteralSpec{
			Value: val,
		},
	}, endPos
}

func lazyFromSeqNode(opts *yamltree.TraverseOpts, node *ast.SequenceNode) (sp manifest2.AnySpec, endPos endPosition, diags parsetypes.Diagnostics) {
	endPos = endPositionFromToken(node.GetToken())
	arr := manifest2.ArraySpec{
		DynamicItems: make([]manifest2.AnySpec, len(node.Values)),
	}

	for i, n := range node.Values {
		if yamltree.IsNullNode(n) {
			endPos = endPositionFromToken(n.GetToken())
			diags = append(diags,
				newErrDiagnosticFromNode(
					opts.FileName, n,
					errors.New("list item cannot be empty"),
				),
			)
			continue
		}

		elemSpec, lastPos, elemDiags := lazyFromNode(opts, n)
		endPos = lastPos
		diags = append(diags, elemDiags...)
		arr.DynamicItems[i] = elemSpec
	}

	//if !diags.HasError() {
	//	return arr.Optimize(), diags
	//}

	return manifest2.AnySpec{
		ArraySpec: &arr,
	}, endPos, diags
}

func lazyFromDictNode(opts *yamltree.TraverseOpts, node *ast.MappingNode) (sp manifest2.AnySpec, endPos endPosition, diags parsetypes.Diagnostics) {
	sp.ObjectSpec = &manifest2.ObjectSpec{
		Values: make(map[string]manifest2.AnySpec, len(node.Values)),
	}

	endPos = endPositionFromToken(node.GetToken())
	for _, n := range node.Values {
		endPos = endPositionFromToken(node.GetToken())
		kn, ok := n.Key.(*ast.StringNode)
		if !ok {
			diags = append(diags,
				newErrDiagnosticFromMapping(
					opts, n,
					errors.New("key should be a string"),
				),
			)
			continue
		}

		if yamltree.IsNullNode(n) {
			diags = append(diags,
				newErrDiagnosticFromMapping(
					opts, n,
					errors.New("node cannot be null"),
				),
			)
			continue
		}

		key := kn.Value
		nodeSpec, tokEndPos, nodeDiags := lazyFromNode(opts, n)
		diags = append(diags, nodeDiags...)
		sp.ObjectSpec.Values[key] = nodeSpec
		endPos = tokEndPos
	}

	//if !diags.HasError() {
	//	sp = sp.Optimize()
	//}

	return sp, endPos, diags
}
