package expr

import (
	"errors"
	"fmt"
)

var (
	UnterminatedExpressionErr = errors.New("unterminated expression")
	EmptyExpressionErr        = errors.New("empty expression")
)

// ExpressionError represents error related to an expression
type ExpressionError struct {
	// ParentRange is range of a parent expression.
	ParentRange Range

	// Range is related statement range.
	Range Range

	// Err is occurred error.
	Err error
}

func newExprError(err error, rng Range) error {
	return &ExpressionError{
		Range: rng,
		Err:   err,
	}
}

func newNestedExprError(err error, rng Range, parRng Range) error {
	return &ExpressionError{
		Range:       rng,
		ParentRange: parRng,
		Err:         err,
	}
}

func (err ExpressionError) Error() string {
	return fmt.Sprintf("%s (at %d:%d)", err.Err, err.Range.StartCol, err.Range.EndCol)
}

func (err ExpressionError) Unwrap() error {
	return err.Err
}
