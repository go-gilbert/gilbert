package profile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func cov2report(cov int) *PackageReport {
	return &PackageReport{
		Coverage: Coverage{Total: 100, Reached: cov},
	}
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

func pkgReportCov(fns map[string]int) *PackageReport {
	out := &PackageReport{
		Functions: make(map[string]*Coverage, len(fns)),
	}

	for c, i := range fns {
		out.Functions[c] = &Coverage{Total: 100, Reached: i}
	}

	return out
}

func TestReport_Sort(t *testing.T) {
	cases := map[string]struct {
		asc    bool
		input  *PackageReport
		expect []string
	}{
		ByName: {
			asc: true,
			input: pkgReportCov(map[string]int{
				"a": 0,
				"c": 10,
				"d": 60,
				"b": 30,
			}),
			expect: []string{"a", "b", "c", "d"},
		},
		ByCoverage: {
			asc: false,
			input: pkgReportCov(map[string]int{
				"a": 0,
				"c": 10,
				"d": 60,
				"b": 30,
			}),
			expect: []string{"d", "b", "c", "a"},
		},
	}

	for sortBy, c := range cases {
		t.Run("sort by "+sortBy, func(t *testing.T) {
			result := c.input.Sort(sortBy, c.asc)
			assert.Equal(t, c.expect, result)
		})
	}
}
