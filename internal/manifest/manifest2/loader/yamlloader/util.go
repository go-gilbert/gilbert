package yamlloader

import (
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

func newErrDiagnosticFromNode(fileName string, node ast.Node, err error) *parsetypes.Diagnostic {
	rng, offset := GetNodeRange(node)
	return &parsetypes.Diagnostic{
		Err:      err,
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: fileName,
		Range:    rng,
		Offset:   offset,
	}
}

func newErrDiagnosticFromMapping(opts *TraverseOpts, node *ast.MappingValueNode, err error) *parsetypes.Diagnostic {
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
