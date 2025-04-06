package expr

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"
)

type tokenExpectation struct {
	Type     TokenType    `json:"type"`
	Content  string       `json:"content"`
	Offset   int          `json:"offset"`
	RawRange *rangeString `json:"range"`
}

func (exp tokenExpectation) Token() *Token {
	var rng parsetypes.Range
	if exp.RawRange != nil {
		rng = exp.RawRange.Range
	}

	return &Token{
		Type:    exp.Type,
		Content: exp.Content,
		Offset:  exp.Offset,
		Range:   rng,
	}
}

type rangeString struct {
	Range parsetypes.Range
}

func (rs *rangeString) UnmarshalText(text []byte) error {
	parts := bytes.SplitN(text, []byte{'-'}, 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid range %q", parts)
	}

	ranges := [2]*parsetypes.Position{
		&rs.Range.Start, &rs.Range.End,
	}

	for i, part := range parts {
		rng, err := posFromBytes(part)
		if err != nil {
			return fmt.Errorf("%w (in %q)", err, text)
		}

		*ranges[i] = rng
	}

	return nil
}

func posFromBytes(b []byte) (r parsetypes.Position, err error) {
	b = bytes.TrimSpace(b)
	parts := bytes.SplitN(b, []byte{':'}, 2)
	if len(parts) != 2 {
		return r, fmt.Errorf("invalid range position %q", b)
	}

	dst := [2]*int{&r.Line, &r.Column}
	for i, chunk := range parts {
		num, err := strconv.Atoi(string(bytes.TrimSpace(chunk)))
		if err != nil {
			return r, err
		}

		*dst[i] = num
	}

	return r, nil
}

const inputsRefPfx = "inputs://"

func resolveContent(caseName, content string) (string, bool) {
	if !strings.HasPrefix(content, inputsRefPfx) {
		return "", false
	}

	p := content[len(inputsRefPfx):]
	if p == "self" {
		p = caseName
	}

	return p, true
}

func trimEOL(src []byte) []byte {
	return trimSuffixes(src, '\r', '\n')
}

func trimSuffixes(src []byte, chars ...byte) []byte {
	for i := len(chars) - 1; i >= 0; i-- {
		if len(src) == 0 {
			return src
		}

		j := len(src) - 1
		if src[j] != chars[i] {
			return src
		}

		src = src[:j]
	}

	return src
}

func lastElem(src []byte) (byte, bool) {
	if len(src) == 0 {
		return 0, false
	}

	return src[len(src)-1], true
}

func loadTxtar(t *testing.T, fname string) fs.FS {
	t.Helper()
	ar, err := txtar.ParseFile(filepath.Join("testdata", fname))
	require.NoError(t, err, "can't open txtar file")

	f, err := txtar.FS(ar)
	require.NoError(t, err, "can't create txtar fs")
	return f
}

func loadYaml[T any](t *testing.T, fname string) (out T) {
	f, err := os.Open(filepath.Join("testdata", fname))
	require.NoError(t, err, "can't open yaml file")

	err = yaml.NewDecoder(f).Decode(&out)
	_ = f.Close()
	require.NoError(t, err)
	return out
}
