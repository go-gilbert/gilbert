package yamlloader

import (
	"fmt"

	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

func newErrDiagnosticFromNode(fileName string, node ast.Node, err error) *parsetypes.Diagnostic {
	rng, offset := yamltree.GetNodeRange(node)
	return &parsetypes.Diagnostic{
		Err:      err,
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: fileName,
		Range:    rng,
		Offset:   offset,
	}
}

func newErrDiagnosticFromMapping(opts *yamltree.TraverseOpts, node *ast.MappingValueNode, err error) *parsetypes.Diagnostic {
	startPos := node.Start.Position
	endPos := node.Key.GetToken().Position

	return &parsetypes.Diagnostic{
		Err:      err,
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: opts.FileName,
		Range: parsetypes.NewRange(
			parsetypes.NewPosition(startPos.Line, startPos.Column),
			parsetypes.NewPosition(endPos.Line, endPos.Column),
		),
		Offset: parsetypes.OffsetRange{
			Start: startPos.Offset,
			End:   endPos.Offset,
		},
	}
}

func endPositionFromToken(tok *token.Token) endPosition {
	return endPosition{
		offset: tok.Position.Offset + len(tok.Value) - 1,
		cursor: parsetypes.Position{
			Line:   tok.Position.Line,
			Column: tok.Position.Column,
		},
	}
}

// copyMapUniq does the same as maps.Copy but avoids duplicates.
func copyMapUniq[V any](dst, src map[string]V) {
	for k, v := range src {
		if _, ok := dst[k]; !ok {
			dst[k] = v
		}
	}
}

// copyMapWithCheck copies src into dst and calls error function on collision.
func copyMapWithCheck[T any](dst, src map[string]T, errFunc func(k string, v T) error) error {
	for k, v := range src {
		if dup, ok := dst[k]; ok {
			return errFunc(k, dup)
		}

		dst[k] = v
	}

	return nil
}

// expressionFromStringNode parses an expression string from a string node.
//
// Last argument is optional reference to a parent multiline node.
// Used to set correct start position of the expression.
func expressionFromStringNode(fileName string, n *ast.StringNode, parent *ast.LiteralNode) (expr.Expression, *parsetypes.Diagnostic) {
	offset, pos := exprPositionFromStrNode(n, parent)
	docInfo := expr.DocumentInfo{
		FileName:      fileName,
		ByteOffset:    offset,
		StartPosition: pos,
	}

	return parseExpr(n.Value, docInfo)
}

func expressionFromAnyNode(fileName string, n ast.Node, v string) (expr.Expression, *parsetypes.Diagnostic) {
	offset, pos := exprPositionFromAnyNode(n)
	docInfo := expr.DocumentInfo{
		FileName:      fileName,
		ByteOffset:    offset,
		StartPosition: pos,
	}

	return parseExpr(v, docInfo)
}

func parseExpr(src string, docInfo expr.DocumentInfo) (expr.Expression, *parsetypes.Diagnostic) {
	parser := expr.NewParser(src, expr.WithDocumentInfo(docInfo))
	v, diag := parser.Parse()
	if diag != nil {
		diag.Err = fmt.Errorf("syntax error in expression: %w", diag.Err)
	}

	return v, diag
}

func exprPositionFromAnyNode(node ast.Node) (int, parsetypes.Position) {
	switch t := node.(type) {
	case *ast.LiteralNode:
		return exprPositionFromStrNode(t.Value, t)
	case *ast.StringNode:
		return exprPositionFromStrNode(t, nil)
	}

	tok := node.GetToken()
	offset := max(tok.Position.Offset, 0)
	pos := parsetypes.NewPosition(tok.Position.Line, tok.Position.Column)

	if tok.Prev != nil && tok.Prev.Position.Line == pos.Line {
		offset = tok.Prev.Position.Offset - 1
		pos = pos.Sub(0, 1)
	}

	return offset, pos
}

func exprPositionFromStrNode(node *ast.StringNode, parent *ast.LiteralNode) (int, parsetypes.Position) {
	if parent != nil {
		// TODO: handle multiline string
	}

	// Expression should start one column before.
	// If string is quoted - use quote position.
	// If string is field value or list item, get position of a space between delimiter (: or -) and a string.
	tok := node.Token
	offset := max(tok.Position.Offset, 0)
	pos := parsetypes.NewPosition(tok.Position.Line, tok.Position.Column)

	if tok.Prev != nil && tok.Prev.Position.Line == pos.Line {
		offset = tok.Prev.Position.Offset - 1
		pos.Column--
		return offset, pos
	}

	// Quoted string?
	if tok.Value == "" || tok.Origin[0] != tok.Value[0] {
		// TODO: seek until string start
		pos.Column--
	}

	return offset, pos
}
