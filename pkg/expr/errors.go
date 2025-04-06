package expr

import (
	"fmt"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type TokenError struct {
	Position parsetypes.Range       `json:"position"`
	Offset   parsetypes.OffsetRange `json:"offset"`
	Err      error                  `json:"err"`
	Note     string                 `json:"note"`
}

func (err *TokenError) Error() string {
	return fmt.Sprintf("%s (at %s - %s)", err.Err, err.Position.Start, err.Position.End)
}

func noteCloseExpr(exprTok string) string {
	return fmt.Sprintf("missing %q", exprTok)
}

func noteRemoveToken(str string) string {
	return fmt.Sprintf("remove unecessary %q", str)
}

func newUnexpectedTokenErr(tok *Token) error {
	return fmt.Errorf("unexpected Token %q", tok.Content)
}
