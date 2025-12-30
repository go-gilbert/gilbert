package parsetypes

import (
	"errors"
	"fmt"
)

type DiagnosticSeverity uint8

func (s DiagnosticSeverity) IsError() bool {
	return s == DiagnosticSeverityError
}

func (s DiagnosticSeverity) String() string {
	switch s {
	case DiagnosticSeverityWarning:
		return "warning"
	case DiagnosticSeverityError:
		return "error"
	default:
		return ""
	}
}

func (s DiagnosticSeverity) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

const (
	DiagnosticSeverityUnknown DiagnosticSeverity = iota
	DiagnosticSeverityError
	DiagnosticSeverityWarning
)

type Diagnostic struct {
	FileName string
	Severity DiagnosticSeverity
	Range    Range
	Offset   OffsetRange
	Note     string
	Err      error
}

func (err *Diagnostic) Unwrap() error {
	return err.Err
}

func (err *Diagnostic) Error() string {
	if err.Range.IsEmpty() {
		return fmt.Sprintf("%s (at %s)", err.Err, err.FileName)
	}

	return fmt.Sprintf("%s (at %s:%d:%d)", err.Err, err.FileName, err.Range.Start.Line, err.Range.Start.Column)
}

type Diagnostics []*Diagnostic

func (diags Diagnostics) Error() string {
	switch len(diags) {
	case 0:
		return "no diagnostics"
	case 1:
		return diags[0].Error()
	default:
		return fmt.Sprintf("%s, and %d other diagnostic(s)", diags[0].Error(), len(diags)-1)
	}
}

func (diags Diagnostics) Append(d *Diagnostic) Diagnostics {
	return append(diags, d)
}

func (diags Diagnostics) Extend(newDiags Diagnostics) Diagnostics {
	return append(diags, newDiags...)
}

func (diags Diagnostics) HasError() bool {
	if len(diags) == 0 {
		return false
	}

	for _, diag := range diags {
		if diag.Severity == DiagnosticSeverityError {
			return true
		}
	}

	return false
}

func HasErrorDiagnostics(diags Diagnostics) bool {
	if len(diags) == 0 {
		return false
	}

	return diags.HasError()
}

// IsDiagnosticsError checks whether error is diagnostics and returns them.
func IsDiagnosticsError(err error) (Diagnostics, bool) {
	var diags Diagnostics
	if errors.As(err, &diags) {
		return diags, true
	}

	diag := new(Diagnostic)
	if errors.As(err, &diag) {
		return Diagnostics{diag}, true
	}

	return nil, false
}

// DiagnosticFromError checks if the error is a Diagnostic and returns it if so.
func DiagnosticFromError(err error) (*Diagnostic, bool) {
	if err == nil {
		return nil, false
	}

	diagnostic := new(Diagnostic)
	if errors.As(err, &diagnostic) {
		return diagnostic, true
	}
	return nil, false
}
