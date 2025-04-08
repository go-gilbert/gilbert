package expr

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type tokenTestCase struct {
	// KeepEOL flag disables end of line char trim.
	// eol trim is enabled by default to mitigate txtar behavior as it always adds eol.
	KeepEOL bool `json:"keepEOL"`

	// UnquoteSource unquotes source from imputs.txtar
	// to parse control characters like \n.
	UnquoteSource bool `json:"unquoteSource"`

	Doc    *DocumentInfo        `json:"doc"`
	Tokens []tokenExpectation   `json:"tokens"`
	Err    *tokenErrExpectation `json:"err"`
}

type tokenTestCases struct {
	// Only is used to test only specific case during debug.
	Only string `json:"only"`

	// Cases is set of expectations and input file in txtar.
	Cases map[string]tokenTestCase
}

func TestTokenizer(t *testing.T) {
	inputsFs := loadTxtar(t, "token-inputs.txtar")
	root := loadYaml[tokenTestCases](t, "token-cases.yml")

	for name, tc := range root.Cases {
		if root.Only != "" && root.Only != name {
			continue
		}

		t.Run(name, func(t *testing.T) {
			wantToks := intoTokens(t, inputsFs, name, tc)
			src, err := fs.ReadFile(inputsFs, name)
			require.NoError(t, err, "input is missing in inputs file")
			if tc.UnquoteSource {
				trimmed, err := strconv.Unquote(`"` + strings.TrimSpace(string(src)) + `"`)
				require.NoError(t, err, "can't unquote input")
				src = []byte(trimmed)
			}

			t.Logf("input: %q", src)
			if !tc.KeepEOL {
				src = trimEOL(src)
			}

			var opts []Option
			if tc.Doc != nil {
				opts = append(opts, WithDocumentInfo(*tc.Doc))
			}

			toker := NewTokenizer(string(src), opts...)
			consumedTokens := make([]*Token, 0, len(tc.Tokens))
			var (
				gotErr  *TokenError
				wantErr = tc.Err.TokenError()
			)

			for tok, err := range toker.IterTokens() {
				if err != nil {
					gotErr = err
					break
				}

				require.NotNil(t, tok, "nil token returned")
				consumedTokens = append(consumedTokens, tok)
			}

			if wantErr != nil {
				require.NotNil(t, gotErr, "expected parser error")
				require.Equal(t, wantErr, gotErr, "parser error doesn't match")
				return
			}

			require.Nil(t, gotErr, "unexpected tokenizer error")
			checkTokenList(t, wantToks, consumedTokens)
		})
	}
}

func intoTokens(t *testing.T, fsys fs.FS, name string, tc tokenTestCase) []*Token {
	t.Helper()
	out := make([]*Token, 0, len(tc.Tokens))
	for _, tokExp := range tc.Tokens {
		tok := tokExp.Token()
		if tokExp.RawContent != "" {
			unescaped, err := strconv.Unquote(`"` + tokExp.RawContent + `"`)
			require.NoErrorf(t, err, "can't unescape: %s", tokExp.RawContent)
			tok.Content = unescaped
			out = append(out, tok)
			continue
		}

		fpath, ok := resolveContent(name, tokExp.Content)
		if ok {
			content, err := fs.ReadFile(fsys, fpath)
			require.NoErrorf(t, err, "can't load referred token content from %q", tok.Content)

			if !tc.KeepEOL {
				content = trimEOL(content)
			}

			tok.Content = string(content)
		}

		out = append(out, tok)
	}

	return out
}

func checkTokenList(t *testing.T, expect, got []*Token) {
	t.Helper()
	require.Equal(
		t, len(expect), len(got),
		"consumed and expected tokens length mismatch",
	)

	if len(expect) == 0 {
		return
	}

	for i, wantTok := range expect {
		gotTok := got[i]
		checkTokenBody(t, i, wantTok, gotTok)

		if i > 0 {
			checkPrevToken(t, i, gotTok, expect[i-1])
		}
	}
}

func checkPrevToken(t *testing.T, pos int, curTok *Token, wantPrevTok *Token) {
	t.Helper()
	if wantPrevTok == nil {
		require.Nil(t, curTok.Prev, "expected prev tok to be nil at %d", pos)
		return
	}

	require.NotNil(t, curTok.Prev, "unexpected nil prev token at %d", pos)
	got := *curTok.Prev
	want := *wantPrevTok
	got.Prev = nil
	want.Prev = got.Prev
	require.Equalf(t, want, got, "mismatch at prev token at %d", pos)
}

func checkTokenBody(t *testing.T, pos int, want, got *Token) {
	t.Helper()
	g := *got
	w := *want
	g.Prev = nil
	w.Prev = g.Prev
	require.Equalf(t, w, g, "mismatch at token %d\n\tA: %q\n\tB: %q", pos, w.Content, g.Content)
}
