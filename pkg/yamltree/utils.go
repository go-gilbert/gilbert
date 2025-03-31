package yamltree

import (
	"errors"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml/ast"
)

func intoDictNode(fi *TraverseOpts, node ast.Node) (*ast.MappingNode, parsetypes.Diagnostics) {
	mn, ok := node.(*ast.MappingNode)
	if !ok {
		return nil, parsetypes.Diagnostics{
			NewErrDiagnosticFromNode(
				fi.FileName, node,
				errors.New("node should be a dictionary"),
			),
		}
	}

	return mn, nil
}

// IsNullNode checks whether node type is null.
func IsNullNode(n ast.Node) bool {
	return n.Type() == ast.NullType
}

// IsPrimitiveNode returns whether node is a string, bool or other scalar type.
func IsPrimitiveNode(n ast.Node) bool {
	switch n.Type() {
	case ast.StringType, ast.IntegerType, ast.FloatType, ast.BoolType, ast.LiteralType:
		return true
	default:
		return false
	}
}

// IsStringNode returns whether node is a string (including multiline).
func IsStringNode(n ast.Node) bool {
	switch n.Type() {
	case ast.StringType, ast.LiteralType:
		return true
	default:
		return false
	}
}

// IntoStringNode attempts to retreive a string node out of untyped node.
func IntoStringNode(n ast.Node) (*ast.StringNode, bool) {
	switch n := n.(type) {
	case *ast.StringNode:
		return n, true
	case *ast.LiteralNode:
		return n.Value, true
	default:
		return nil, false
	}
}

// GetNodeRange returns node document range
func GetNodeRange(node ast.Node) (parsetypes.Range, parsetypes.OffsetRange) {
	startTok := node.GetToken()
	startPos := startTok.Position
	rng := parsetypes.NewRange(
		parsetypes.NewPosition(startPos.Line, startPos.Column),
		parsetypes.NewPosition(startPos.Line, startPos.Column),
	)

	offset := parsetypes.OffsetRange{
		Start: startPos.Offset + startPos.IndentNum,
		End:   startPos.Offset + startPos.IndentNum,
	}

	switch n := node.(type) {
	case *ast.StringNode,
		*ast.IntegerNode,
		*ast.FloatNode,
		*ast.BoolNode:
		strlen := len(startTok.Value)
		rng.End.Column += strlen - 1
		offset.End += strlen
	case *ast.LiteralNode:
		rng, offset = getLiteralRange(n)
	}

	return rng, offset
}
