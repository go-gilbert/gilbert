package parsetypes

import (
	"fmt"
)

type SeverityError interface {
	Severity() DiagnosticSeverity
}

// AnnotatedError is error that contains additional note for a user with a detail.
type AnnotatedError struct {
	err      error
	note     string
	severity DiagnosticSeverity
}

func NewAnnotatedError(err error, format string, args ...any) *AnnotatedError {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	return &AnnotatedError{
		err:      err,
		note:     msg,
		severity: DiagnosticSeverityError,
	}
}

func (err *AnnotatedError) Severity() DiagnosticSeverity {
	return err.severity
}

func (err *AnnotatedError) AsWarning() *AnnotatedError {
	err.severity = DiagnosticSeverityWarning
	return err
}

func (err *AnnotatedError) Note() string {
	return err.note
}

func (err *AnnotatedError) Error() string {
	return err.err.Error()
}

func (err *AnnotatedError) Unwrap() error {
	return err.err
}

type WarningError struct {
	Err error
}

func (err WarningError) Severity() DiagnosticSeverity {
	return DiagnosticSeverityWarning
}

func (err WarningError) Error() string {
	return err.Err.Error()
}

func (err WarningError) Unwrap() error {
	return err.Err
}

// WrapAsWarning marks error as warning to set a corresponding severity for diagnostic.
func WrapAsWarning(err error) WarningError {
	return WarningError{
		Err: err,
	}
}

// SeverityFromError returns diagnostic severity for an error.
//
// Returns [DiagnosticSeverityWarning] if error implements [SeverityError].
func SeverityFromError(err error) DiagnosticSeverity {
	if s, ok := err.(SeverityError); ok {
		return s.Severity()
	}

	return DiagnosticSeverityError
}
