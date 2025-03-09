package parsetypes

import "fmt"

type DiagnosticSeverity uint8

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
	Severity DiagnosticSeverity
	FileName string
	Range    Range
	Offset   OffsetRange
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
