package profile

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func cov2report(cov int) *PackageReport {
	return &PackageReport{
		Coverage: Coverage{Total: 100, Reached: cov},
	}
}

func strReverse(in []string) []string {
	out := make([]string, 0, len(in))
	for k := range in {
		out = append(out, in[k])
	}

	return out
}

func TestPackages_Sort(t *testing.T) {
	cases := map[string]struct {
		asc    bool
		input  Packages
		expect []string
	}{
		ByName: {
			asc: true,
			input: Packages{
				"github.com/go-gilbert/gilbert/bcc/aaa": nil,
				"github.com/go-gilbert/gilbert/fca":     nil,
				"github.com/go-gilbert/gilbert/bcc":     nil,
				"github.com/go-gilbert/gilbert/abc":     nil,
			},
			expect: []string{
				"github.com/go-gilbert/gilbert/abc",
				"github.com/go-gilbert/gilbert/bcc",
				"github.com/go-gilbert/gilbert/bcc/aaa",
				"github.com/go-gilbert/gilbert/fca",
			},
		},
		ByCoverage: {
			asc: false,
			input: Packages{
				"c": cov2report(30),
				"d": cov2report(25),
				"b": cov2report(78),
				"e": cov2report(0),
				"a": cov2report(100),
			},
			expect: []string{"a", "b", "c", "d", "e"},
		},
	}

	for sortBy, c := range cases {
		t.Run("sort by "+sortBy+"(asc)", func(t *testing.T) {
			result := c.input.Sort(sortBy, c.asc)
			assert.Equal(t, c.expect, result)
		})
	}
}
