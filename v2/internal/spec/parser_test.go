package spec

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"
)

func TestParse(t *testing.T) {
	fname, err := filepath.Abs("testdata/sample.hcl")
	require.NoError(t, err)
	//fname := "testdata/simple.hcl"
	data, err := os.ReadFile(fname)
	require.NoError(t, err)

	rootCtx := NewRootContext()
	rootCtx.Runner = RunnerSpec{
		Version:   "v2.0.0",
		Build:     "1",
		CommitSHA: "n/a",
	}

	parser := NewParser(NewRootContext(), ProjectSpec{
		FileName:         fname,
		WorkingDirectory: filepath.Dir(fname),
	})
	m, err := parser.Parse(data)
	if err == nil {
		spew.Dump(m)
		return
	}

	errs, ok := err.(hcl.Diagnostics)
	if !ok {
		t.Fatal(err)
	}

	if len(errs) > 0 {
		t.Logf("%d Errors during parsing", len(errs))
		for i, err := range errs {
			rng := oneOfNotNil(err.Subject, err.Context)
			src := readRange(data, rng)
			t.Errorf("E:%d - %s:\n%s", i, err,
				src)
		}
		return
	}
}

func oneOfNotNil[T any](a, b *T) *T {
	if a != nil {
		return a
	}

	return b
}

func readRange(data []byte, rng *hcl.Range) []byte {
	start := rng.Start.Byte
	end := rng.End.Byte
	if end-start <= 2 {
		return readUntilLineBreak(data[start:])
	}
	return data[start:end]
}

func readUntilLineBreak(data []byte) []byte {
	offset := 0
	for {
		if len(data) <= offset {
			return data
		}
		char, size := utf8.DecodeRune(data[offset:])
		offset += size
		switch char {
		case utf8.RuneError:
			return []byte("RUNE ERROR")
		case '\n':
			return data[:offset]
		}
	}
}
