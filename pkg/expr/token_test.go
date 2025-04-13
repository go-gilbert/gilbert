package expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToken_EndOffset(t *testing.T) {
	tok := &Token{
		Offset: 10,
	}
	require.Equal(t, tok.Offset, tok.EndOffset())

	tok.Content = "12345"
	require.Equal(t, tok.Offset+len(tok.Content)-1, tok.EndOffset())
}

func TestTokenType_GetPair(t *testing.T) {
	m := map[TokenType]TokenType{
		TokenTypeShellStart: TokenTypeShellEnd,
		TokenTypeEvalStart:  TokenTypeEvalEnd,
	}
	misc := []TokenType{
		TokenTypeString,
		TokenTypeEOL,
	}

	for a, b := range m {
		require.Equal(t, a.GetPair(), b)
		require.Equal(t, b.GetPair(), a)
	}

	for _, tok := range misc {
		require.Equal(t, tok.GetPair(), tok)
	}
}

func TestTokenType_IsPairOf(t *testing.T) {
	m := map[TokenType]TokenType{
		TokenTypeShellStart: TokenTypeShellEnd,
		TokenTypeEvalStart:  TokenTypeEvalEnd,
	}
	misc := []TokenType{
		TokenTypeString,
		TokenTypeEOL,
	}

	for a, b := range m {
		require.True(t, a.IsPairOf(b))
		require.True(t, b.IsPairOf(a))

		for _, tok := range misc {
			require.False(t, a.IsPairOf(tok))
			require.False(t, b.IsPairOf(tok))
		}
	}
}

func TestTokenType_Text(t *testing.T) {
	m := map[TokenType]string{
		TokenTypeShellStart: string(exprPrefix) + shellStartTok,
		TokenTypeShellEnd:   shellEndTok,
		TokenTypeEvalStart:  string(exprPrefix) + evalStartTok,
		TokenTypeEvalEnd:    evalEndTok,
		TokenTypeEOL:        "\n",
		TokenTypeString:     "",
		TokenTypeEmpty:      "",
	}

	for input, want := range m {
		require.Equal(t, input.Text(), want)
	}
}

func TestTokenType_String(t *testing.T) {
	m := map[TokenType]string{
		TokenTypeShellStart: "ShellStart",
		TokenTypeShellEnd:   "ShellEnd",
		TokenTypeEvalStart:  "EvalStart",
		TokenTypeEvalEnd:    "EvalEnd",
		TokenTypeString:     "String",
		TokenTypeEOL:        "EOL",
		TokenTypeEmpty:      "",
	}

	for input, want := range m {
		require.Equal(t, input.String(), want)
	}
}
