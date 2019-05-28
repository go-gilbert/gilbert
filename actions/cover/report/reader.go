/*
package lines parses output of "go test" tool, formats and provides information about uncovered packages and failed tests
*/
package report

import (
	"fmt"
	"strings"
)

var ident = 2

type Formatter struct {
	lines Lines
}

// Write implements io.Writer
func (a *Formatter) Write(p []byte) (n int, err error) {
	return len(p), a.lines.AppendData(p)
}

// UncoveredPackages returns pre-formatted lines about uncovered packages
func (a *Formatter) UncoveredPackages() (string, int) {
	pkgs := a.lines.SkippedPackages()
	pkgCount := len(pkgs)
	b := strings.Builder{}
	for _, s := range pkgs {
		// format: "  - package_name"
		b.WriteString(wrapListItem(1, "- "+s))
	}

	return b.String(), pkgCount
}

// FailedTests returns pre-formatted lines about failed tests
func (a *Formatter) FailedTests() (string, int) {
	failures := a.lines.Failed()
	if len(failures) == 0 {
		return "", 0
	}

	failedCount := 0
	b := &strings.Builder{}
	for pkg, tests := range failures {
		/*
			Format:
				package "github.com/name/pkg":
					- Test:
						main.go: Test error
		*/
		b.WriteString(wrapListItem(1, fmt.Sprintf(`Package "%s":`, pkg)))

		for test, errors := range tests {
			b.WriteString(wrapListItem(2, test))
			outputLinesToList(b, errors)
			failedCount++
		}
	}

	return b.String(), failedCount
}

func outputLinesToList(w *strings.Builder, lines []string) {
	for _, l := range lines {
		w.WriteString(wrapListItem(3, l+"\n"))
	}
}

// NewReportFormatter returns new "go test" tool report formatter
//
// It reads and formats output from stdout of "go test" command
func NewReportFormatter() *Formatter {
	return &Formatter{
		lines: make(Lines, 0),
	}
}

func wrapListItem(level int, str string) string {
	return strings.Repeat(" ", ident*level) + str + "\n"
}
