package yamltree

import (
	"strings"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml/ast"
)

func NewErrDiagnosticFromNode(fileName string, node ast.Node, err error) *parsetypes.Diagnostic {
	rng, offset := GetNodeRange(node)

	return &parsetypes.Diagnostic{
		Err:      err,
		Severity: parsetypes.SeverityFromError(err),
		FileName: fileName,
		Range:    rng,
		Offset:   offset,
	}
}

func getLiteralRange(n *ast.LiteralNode) (parsetypes.Range, parsetypes.OffsetRange) {
	// Determine length of literal prefix, including new line and spaces.
	// (necessary as value is trimmed in AST).
	endOffset := n.Start.Position.Offset + len(n.Value.Token.Origin)
	headerStartPos := strings.Index(n.Start.Origin, n.Start.Value)
	if headerStartPos != -1 {
		endOffset += len(n.Start.Origin) - headerStartPos
	}

	// Determine lines count
	linesCount := 0
	charCount := 0
	lastChunkLen := 0
	for _, char := range n.Value.Token.Origin {
		switch char {
		case '\r', '\n':
			if charCount > 0 {
				linesCount++
				lastChunkLen = charCount
			}
			charCount = 0
		default:
			charCount++
		}
	}

	rng := parsetypes.NewRange(
		parsetypes.NewPosition(n.Start.Position.Line, n.Start.Position.Column),
		parsetypes.NewPosition(n.Start.Position.Line+linesCount, max(lastChunkLen, 1)),
	)

	offset := parsetypes.OffsetRange{
		Start: n.Start.Position.Offset,
		End:   endOffset,
	}

	return rng, offset
}

func newErrDiagnosticFromMapping(opts *TraverseOpts, node *ast.MappingValueNode, err error) *parsetypes.Diagnostic {
	startPos := node.Key.GetToken().Position
	endPos := node.Start.Position

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
