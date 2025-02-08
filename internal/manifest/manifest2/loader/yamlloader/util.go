package yamlloader

import (
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

func newErrDiagnosticFromNode(fileName string, node ast.Node, err error) *parsetypes.Diagnostic {
	tok := node.GetToken()
	startPos := tok.Prev.Position
	endPos := tok.Position
	return &parsetypes.Diagnostic{
		Err:      err,
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: fileName,
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

func newErrDiagnosticFromMapping(fi yamltree.FileInfo, node *ast.MappingValueNode, err error) *parsetypes.Diagnostic {
	startPos := node.Start.Position
	endPos := node.Key.GetToken().Position

	return &parsetypes.Diagnostic{
		Err:      err,
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: fi.FileName,
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

func newSimpleErrorDiagnostic(fileName string, err error) *parsetypes.Diagnostic {
	return &parsetypes.Diagnostic{
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: fileName,
		Err:      err,
	}
}
