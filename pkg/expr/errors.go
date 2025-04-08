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

func newUnexpectedTokenErr(tok *Token) error {
	return fmt.Errorf("unexpected token %q", tok.Content)
}
