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
	TokenTypeEOL
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
	case "EOL":
		return TokenTypeEOL
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
	case TokenTypeEOL:
		return "EOL"
	default:
		return ""
	}
}

func (t TokenType) Text() string {
	switch t {
	case TokenTypeEvalStart:
		return string(exprPrefix) + evalStartTok
	case TokenTypeEvalEnd:
		return evalEndTok
	case TokenTypeShellStart:
		return string(exprPrefix) + shellStartTok
	case TokenTypeShellEnd:
		return shellEndTok
	case TokenTypeEOL:
		return "\n"
	default:
		return ""
	}
}

func (t TokenType) GetPair() TokenType {
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

func (t TokenType) IsClosedBy(tok TokenType) bool {
	switch t {
	case TokenTypeEvalStart:
		return tok == TokenTypeEvalEnd
	case TokenTypeShellStart:
		return tok == TokenTypeShellEnd
	default:
		return false
	}
}

func (t TokenType) IsOpenedBy(tok TokenType) bool {
	switch t {
	case TokenTypeEvalEnd:
		return tok == TokenTypeEvalStart
	case TokenTypeShellEnd:
		return tok == TokenTypeShellStart
	default:
		return false
	}
}

func (t TokenType) IsPairOf(tok TokenType) bool {
	return t.IsClosedBy(tok) || t.IsOpenedBy(tok)
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
