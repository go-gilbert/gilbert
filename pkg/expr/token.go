package expr

import (
	"fmt"
	"strings"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

const (
	exprPrefix = '$'

	shellStartTok = "("
	shellEndTok   = ")"
	evalStartTok  = "{{"
	evalEndTok    = "}}"
)

type TokenType uint

const (
	TokenTypeEmpty TokenType = iota
	TokenTypeString
	TokenTypeShellStart
	TokenTypeShellEnd
	TokenTypeEvalStart
	TokenTypeEvalEnd
)

func TokenTypeFromString(str string) TokenType {
	switch str {
	case "EvalStart":
		return TokenTypeEvalStart
	case "EvalEnd":
		return TokenTypeEvalEnd
	case "ShellStart":
		return TokenTypeShellStart
	case "ShellEnd":
		return TokenTypeShellEnd
	case "String":
		return TokenTypeString
	default:
		return TokenTypeEmpty
	}
}

func (t *TokenType) UnmarshalText(text []byte) error {
	str := string(text)
	*t = TokenTypeFromString(str)

	if *t == TokenTypeEmpty && str != "" && str != "Empty" {
		return fmt.Errorf("invalid token type: %s", str)
	}

	return nil
}

func (t TokenType) MarshalText() ([]byte, error) {
	s := t.String()
	if s == "" {
		return nil, fmt.Errorf("unknown token type: %d", t)
	}
	return []byte(s), nil
}

func (t TokenType) String() string {
	switch t {
	case TokenTypeEvalStart:
		return "EvalStart"
	case TokenTypeEvalEnd:
		return "EvalEnd"
	case TokenTypeShellStart:
		return "ShellStart"
	case TokenTypeShellEnd:
		return "ShellEnd"
	case TokenTypeString:
		return "String"
	default:
		return ""
	}
}

func (t TokenType) getPair() TokenType {
	switch t {
	case TokenTypeEvalStart:
		return TokenTypeEvalEnd
	case TokenTypeEvalEnd:
		return TokenTypeEvalStart
	case TokenTypeShellStart:
		return TokenTypeShellEnd
	case TokenTypeShellEnd:
		return TokenTypeShellStart
	default:
		return t
	}
}

func (t TokenType) isClosedBy(tok TokenType) bool {
	switch t {
	case TokenTypeEvalStart:
		return tok == TokenTypeEvalEnd
	case TokenTypeShellStart:
		return tok == TokenTypeShellEnd
	default:
		return false
	}
}

func (t TokenType) isOpenedBy(tok TokenType) bool {
	switch t {
	case TokenTypeEvalEnd:
		return tok == TokenTypeEvalStart
	case TokenTypeShellEnd:
		return tok == TokenTypeShellStart
	default:
		return false
	}
}

func (t TokenType) isPairOf(tok TokenType) bool {
	return t.isClosedBy(tok) || t.isOpenedBy(tok)
}

func (t TokenType) isCloseToken() bool {
	switch t {
	case TokenTypeEvalEnd, TokenTypeShellEnd:
		return true
	}

	return false
}

func (t TokenType) isOpenToken() bool {
	switch t {
	case TokenTypeShellStart, TokenTypeEvalStart:
		return true
	}

	return false
}

type Token struct {
	Prev    *Token           `json:"prev"`
	Type    TokenType        `json:"type"`
	Content string           `json:"content"`
	Offset  int              `json:"offset"`
	Range   parsetypes.Range `json:"range"`
}

// EndOffset returns token last content index position.
func (t *Token) EndOffset() int {
	if t.Content == "" {
		return t.Offset
	}

	return t.Offset + len(t.Content) - 1
}

// hasOpenTokenPrefix checks if there is token open clause.
func hasOpenTokenPrefix(str string) (TokenType, bool) {
	switch {
	case strings.HasPrefix(str, evalEndTok):
		return TokenTypeEvalEnd, true
	case strings.HasPrefix(str, shellEndTok):
		return TokenTypeShellEnd, true
	default:
		return TokenTypeEmpty, false
	}
}

// hasTokenClosePrefix checks if string starts with shell or eval expression close Token.
//
// Returns Token type and its contents.
// Returns TokenTypeEmpty if nothing found.
func hasTokenClosePrefix(str string) (TokenType, string) {
	switch {
	case strings.HasPrefix(str, evalEndTok):
		return TokenTypeEvalEnd, evalEndTok
	case strings.HasPrefix(str, shellEndTok):
		return TokenTypeShellEnd, shellEndTok
	default:
		return TokenTypeEmpty, ""
	}
}

func tokenToString(t TokenType) string {
	switch t {
	case TokenTypeEvalStart:
		return string(exprPrefix) + evalStartTok
	case TokenTypeEvalEnd:
		return evalEndTok
	case TokenTypeShellStart:
		return string(exprPrefix) + shellStartTok
	case TokenTypeShellEnd:
		return shellEndTok
	default:
		return ""
	}
}
