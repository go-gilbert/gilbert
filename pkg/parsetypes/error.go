package parsetypes

import (
	"errors"
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

	note string
}

func (err *WarningError) Severity() DiagnosticSeverity {
	return DiagnosticSeverityWarning
}

func (err *WarningError) Error() string {
	return err.Err.Error()
}

func (err *WarningError) Unwrap() error {
	return err.Err
}

func (err *WarningError) Note() string {
	return err.note
}

// WithNote adds formatted note to a diagnostic error.
func (err *WarningError) WithNote(format string, args ...any) *WarningError {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	err.note = msg
	return err
}

// WrapAsWarning marks error as warning to set a corresponding severity for diagnostic.
func WrapAsWarning(err error) *WarningError {
	return &WarningError{
		Err: err,
	}
}

// Warningf returns a warning error with following message.
func Warningf(format string, args ...any) *WarningError {
	if len(args) == 0 {
		return &WarningError{
			Err: errors.New(format),
		}
	}

	return &WarningError{
		Err: fmt.Errorf(format, args...),
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

type ErrorWithNote interface {
	Note() string
}

// NoteFromError returns diagnostic note from an error, if error implements [ErrorWithNote] interface.
func NoteFromError(err error) string {
	if v, ok := err.(ErrorWithNote); ok {
		return v.Note()
	}

	return ""
}
