package expr2

import (
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
	TokenTypeExprStart
	TokenTypeExprEnd
)

func (t TokenType) getPair() TokenType {
	switch t {
	case TokenTypeExprStart:
		return TokenTypeExprEnd
	case TokenTypeExprEnd:
		return TokenTypeExprStart
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
	case TokenTypeExprStart:
		return tok == TokenTypeExprEnd
	case TokenTypeShellStart:
		return tok == TokenTypeShellEnd
	default:
		return false
	}
}

func (t TokenType) isOpenedBy(tok TokenType) bool {
	switch t {
	case TokenTypeExprEnd:
		return tok == TokenTypeExprStart
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
	case TokenTypeExprEnd, TokenTypeShellEnd:
		return true
	}

	return false
}

func (t TokenType) isOpenToken() bool {
	switch t {
	case TokenTypeShellStart, TokenTypeExprStart:
		return true
	}

	return false
}

type Token struct {
	prev    *Token
	typ     TokenType
	content string
	offset  int
	rng     parsetypes.Range
}

// hasTokenClosePrefix checks if string starts with shell or eval expression close Token.
//
// Returns Token type and its contents.
// Returns TokenTypeEmpty if nothing found.
func hasTokenClosePrefix(str string) (TokenType, string) {
	switch {
	case strings.HasPrefix(str, evalEndTok):
		return TokenTypeExprEnd, evalEndTok
	case strings.HasPrefix(str, shellEndTok):
		return TokenTypeShellEnd, shellEndTok
	default:
		return TokenTypeEmpty, ""
	}
}

func tokenToString(t TokenType) string {
	switch t {
	case TokenTypeExprStart:
		return string(exprPrefix) + evalStartTok
	case TokenTypeExprEnd:
		return evalEndTok
	case TokenTypeShellStart:
		return string(exprPrefix) + shellStartTok
	case TokenTypeShellEnd:
		return shellEndTok
	default:
		return ""
	}
}
