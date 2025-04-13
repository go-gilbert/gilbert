package yamltree

import (
	"errors"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/token"
)

func tryMapError[TErr error](err error, mapFn func(TErr) *parsetypes.Diagnostic) (*parsetypes.Diagnostic, bool) {
	var typedErr TErr
	if !errors.As(err, &typedErr) {
		return nil, false
	}

	return mapFn(typedErr), true
}

func errorToDiagnostics(fileName string, err error) parsetypes.Diagnostics {
	diags, ok := tryMapError(err, func(serr *yaml.SyntaxError) *parsetypes.Diagnostic {
		return errDiagnosticFromToken(fileName, serr.Message, serr.Token)
	})
	if ok {
		return parsetypes.Diagnostics{diags}
	}

	diags, ok = tryMapError(err, func(serr *yaml.DuplicateKeyError) *parsetypes.Diagnostic {
		return errDiagnosticFromToken(fileName, serr.Message, serr.Token)
	})
	if ok {
		return parsetypes.Diagnostics{diags}
	}

	diags, ok = tryMapError(err, func(serr *yaml.UnexpectedNodeTypeError) *parsetypes.Diagnostic {
		return errDiagnosticFromToken(fileName, serr.Error(), serr.Token)
	})
	if ok {
		return parsetypes.Diagnostics{diags}
	}

	if diag, ok := parsetypes.DiagnosticFromError(err); ok {
		return parsetypes.Diagnostics{diag}
	}

	return parsetypes.Diagnostics{
		&parsetypes.Diagnostic{
			Severity: parsetypes.DiagnosticSeverityError,
			FileName: fileName,
			Err:      err,
		},
	}
}

func errDiagnosticFromToken(fileName, message string, tok *token.Token) *parsetypes.Diagnostic {
	pos := tok.Position
	return &parsetypes.Diagnostic{
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: fileName,
		Range: parsetypes.NewRange(
			parsetypes.NewPosition(pos.Line, pos.Column),
			parsetypes.NewPosition(pos.Line, pos.Column+len(tok.Value)-1),
		),
		Offset: parsetypes.OffsetRange{
			Start: pos.Offset,
			End:   pos.Offset + 1,
		},
		Err: errors.New(message),
	}
}
