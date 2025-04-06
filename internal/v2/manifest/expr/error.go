package expr

import (
	"errors"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

var (
	ErrBadToken               = errors.New("invalid token")
	ErrNestedShellExpression  = errors.New("shell expression cannot contain another shell expression")
	ErrUnterminatedExpression = errors.New("unterminated expression")
	ErrEmptyExpression        = errors.New("empty expression")
)

// ExpressionError represents error related to an expression
type ExpressionError struct {
	// ParentRange is range of a parent expression.
	ParentLocation parsetypes.Location

	// Range is related statement range.
	Location parsetypes.Location

	// Err is occurred error.
	Err error
}

func newExprError(err error, pos documentPos) *ExpressionError {
	return &ExpressionError{
		Location: pos.location(),
		Err:      err,
	}
}

func newNestedExprError(err error, pos, parentPos documentPos) *ExpressionError {
	return &ExpressionError{
		Location:       pos.location(),
		ParentLocation: parentPos.location(),
		Err:            err,
	}
}

func (err ExpressionError) Error() string {
	return err.Err.Error()
}

func (err ExpressionError) Unwrap() error {
	return err.Err
}

func isUnterminatedErr(err *ExpressionError) bool {
	return errors.Is(err.Err, ErrUnterminatedExpression)
}

func convertEvalError(err error) {

}
