package scope

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVarsScan(t *testing.T) {
	c := &Scope{
		Variables: Vars{
			"foo": "bar",
		},
	}

	c.processor = NewExpressionProcessor(c)
	input := "foo{{foo}}"
	assert.NoError(t, c.Scan(&input))
	assert.Equal(t, "foobar", input)
}

func TestVarsExtract(t *testing.T) {
	c := &Scope{
		Globals: Vars{
			"GOROOT": "/usr/local/go",
			"GOPATH": "/home/root/go",
		},
		Variables: Vars{
			"package": "github.com/x1unix/gorn",
			"nested":  "{{GOPATH}}/foo",
		},
	}

	c.processor = NewExpressionProcessor(c)

	cases := map[string]struct {
		input       string
		shouldError bool
		expString   string
		trimResult  bool
	}{
		"should extract valid variables": {
			input:     "{{GOROOT}}/src/{{ package }}",
			expString: fmt.Sprintf("%s/src/%s", c.Globals["GOROOT"], c.Variables["package"]),
		},
		"another extract test": {
			input:     "/var/lib/{{GOPATH}}/foo",
			expString: fmt.Sprintf("/var/lib/%s/foo", c.Globals["GOPATH"]),
		},
		"should include nested variables in local variable": {
			input:     "/var/{{nested}}/bar",
			expString: fmt.Sprintf("/var/%s/foo/bar", c.Globals["GOPATH"]),
		},
		"should fail on undefined variable": {
			input:       "/foo/{{ bar }}/baz",
			shouldError: true,
			expString:   "variable 'bar' is undefined",
		},
		"should expand shell expression": {
			trimResult: true,
			input:      "foo {% echo bar %}",
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
					tt.Fatalf("returned error - %v", err)
					return
				}

				if !strings.Contains(err.Error(), test.expString) {
					tt.Fatalf("bad error message\n\nWant: %s\nGot: %s", test.expString, err.Error())
				}
				return
			}

			if test.trimResult {
				got = strings.TrimSpace(got)
			}
			if got != test.expString {
				tt.Fatalf("result mismatch\n\nWant: %s\nGot: %s", test.expString, got)
			}
		})
	}

}
