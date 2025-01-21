package scope

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/manifest/expr"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestVarsScan(t *testing.T) {
	c := &Scope{
		Variables: manifest.Vars{
			"foo": "bar",
		},
	}

	c.parser = expr.SpecV2Parser{}
	input := "foo${foo}"
	assert.NoError(t, c.Scan(&input))
	assert.Equal(t, "foobar", input)
}

func TestVarsExtract(t *testing.T) {
	c := &Scope{
		Globals: manifest.Vars{
			"GOROOT": "/usr/local/go",
			"GOPATH": "/home/root/go",
		},
		Variables: manifest.Vars{
			"package": "github.com/go-gilbert/gorn",
			"nested":  "${GOPATH}/foo",
		},
	}

	c.parser = expr.SpecV2Parser{}

	cases := map[string]struct {
		input       string
		shouldError bool
		expString   string
		trimResult  bool
	}{
		"should extract valid variables": {
			input:     "${GOROOT}/src/${ package }",
			expString: fmt.Sprintf("%s/src/%s", c.Globals["GOROOT"], c.Variables["package"]),
		},
		"another extract test": {
			input:     "/var/lib/${GOPATH}/foo",
			expString: fmt.Sprintf("/var/lib/%s/foo", c.Globals["GOPATH"]),
		},
		"should include nested variables in local variable": {
			input:     "/var/${nested}/bar",
			expString: fmt.Sprintf("/var/%s/foo/bar", c.Globals["GOPATH"]),
		},
		"should fail on undefined variable": {
			input:       "/foo/${ bar }/baz",
			shouldError: true,
			expString:   `"bar" is not defined`,
		},
		"should expand shell expression": {
			trimResult: true,
			input:      "foo $( echo bar )",
			expString:  "foo bar",
		},
		"should ignore non-complete statement": {
			input:     "/a$b",
			expString: "/a$b",
		},
	}

	for name, test := range cases {
		t.Run(name, func(tt *testing.T) {
			got, err := c.ExpandVariables(test.input)
			if err != nil {
				if !test.shouldError {
					require.NoError(tt, err)
					return
				}

				require.Error(tt, err)
				require.Contains(tt, err.Error(), test.expString)
				return
			}

			if test.trimResult {
				got = strings.TrimSpace(got)
			}

			require.Equal(tt, test.expString, got)
		})
	}

}
