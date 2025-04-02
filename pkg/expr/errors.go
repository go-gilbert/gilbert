package expr2

import (
	"errors"
	"fmt"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

var (
	errDoubleShellExpr = errors.New("shell expression cannot contain another shell expression")

	ErrUnterminatedShellExpr = errors.New("unterminated shell expression")
	ErrUnterminatedEvalExpr  = errors.New("unterminated eval expression")
)

// ${} $(${})
// $() $(${}
// $(${$(

type TokenError struct {
	Position parsetypes.Range
	Offset   parsetypes.OffsetRange
	Err      error
	Note     string
}

func (err *TokenError) Error() string {
	return fmt.Sprintf("%s (at %s)", err, err.Position)
}

func noteCloseExpr(exprTok string) string {
	return fmt.Sprintf("missing %q", exprTok)
}

func noteRemoveToken(str string) string {
	return fmt.Sprintf("remove unecessary %q", str)
}

func newUnexpectedTokenErr(tok *Token) error {
	return fmt.Errorf("unexpected Token %q", tok.content)
}
