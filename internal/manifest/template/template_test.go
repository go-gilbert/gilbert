package template

import (
	"fmt"
	"testing"

	"github.com/go-gilbert/gilbert/internal/support/test"
	"github.com/stretchr/testify/assert"
)

func TestCompileManifest(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		err      string
	}{
		{
			input:    `{{ shell "echo foo,bar" | split "," | yaml }}`,
			expected: `["foo","bar"]`,
		},
		{
			input: `{{`,
			err:   "unclosed action",
		},
		{
			input: `{{ shell "blablabla" }}`,
			err:   "returned error",
		},
		{
			input:    `{{ slice "foo" "bar" | yaml }}`,
			expected: `["foo","bar"]`,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			out, err := CompileManifest([]byte(c.input))
			if c.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, string(out))
				return
			}

			test.AssertErrorContains(t, err, c.err)
		})
	}
}
