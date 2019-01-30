package env

import (
	"fmt"
	"strings"
	"testing"
)

func TestVarsExtract(t *testing.T) {
	c := Context{
		Globals: Vars{
			"GOROOT": "/usr/local/go",
			"GOPATH": "/home/root/go",
		},
		Variables: Vars{
			"package": "github.com/x1unix/gorn",
			"nested":  "$(GOPATH)/foo",
		},
	}

	cases := map[string]struct {
		input       string
		shouldError bool
		expString   string
	}{
		"should extract valid variables": {
			input:     "$(GOROOT)/src/$(package)",
			expString: fmt.Sprintf("%s/src/%s", c.Globals["GOROOT"], c.Variables["package"]),
		},
		"another extract test": {
			input:     "/var/lib/$(GOPATH)/foo",
			expString: fmt.Sprintf("/var/lib/%s/foo", c.Globals["GOPATH"]),
		},
		"should include nested variables in local variable": {
			input:     "/var/$(nested)/bar",
			expString: fmt.Sprintf("/var/%s/foo/bar", c.Globals["GOPATH"]),
		},
		"should fail on undefined variable": {
			input:       "/foo/$(bar)/baz",
			shouldError: true,
			expString:   "variable 'bar' is undefined",
		},
		"should fail on unterminated statement": {
			input:       "/f$(boobaabeer",
			shouldError: true,
			expString:   "expression is not finished",
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

			if got != test.expString {
				tt.Fatalf("result mismatch\n\nWant: %s\nGot: %s", test.expString, got)
			}
		})
	}

}
