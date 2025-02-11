package yamlloader

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
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

func getDefaultValueVisitor(_ context.Context, node ast.Node, def *manifest2.InputDefinition) (ValueVisitor[any], error) {
	if !IsPrimitiveNode(node) {
		return nil, fmt.Errorf("expected %s but got %s", def.Type, node.Type())
	}

	switch def.Type {
	case manifest2.ValueTypeString:
		return Transform[string, any](String(), func(ctx context.Context, node ast.Node, s string) (any, error) {
			switch def.Format {
			case manifest2.ValueFormatDate:
				format := def.DateFormat
				if format == "" {
					format = manifest2.DefaultDateFormat
				}

				return time.Parse(format, s)
			case manifest2.ValueFormatDuration:
				return time.ParseDuration(s)
			case manifest2.ValueFormatURL:
				_, err := url.Parse(s)
				return s, err
			}
			return s, nil
		}), nil
	case manifest2.ValueTypeInt:
		return Transform[int64, any](Int[int64](), intoAny), nil
	case manifest2.ValueTypeFloat:
		return Transform[float64, any](Float[float64](), intoAny), nil
	case manifest2.ValueTypeBool:
		return Transform[bool, any](Bool(), intoAny), nil
	default:
		// TODO: support lists?
		return nil, fmt.Errorf("default values for input of type %q are not supported", def.Type)
	}
}

func intoAny[T any](_ context.Context, _ ast.Node, s T) (any, error) {
	return s, nil
}
