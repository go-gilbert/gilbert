package parsetypes

import "fmt"

// AnnotatedError is error that contains additional note for a user with a detail.
type AnnotatedError struct {
	err  error
	note string
}

func NewAnnotatedError(err error, format string, args ...any) *AnnotatedError {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	return &AnnotatedError{
		err:  err,
		note: msg,
	}
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
