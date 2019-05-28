package report

import (
	"encoding/json"
	"strings"
)

const (
	actionOutput = "output"
	actionPass   = "pass"
	actionSkip   = "skip"
	actionFail   = "fail"
	actionRun    = "run"
)

// ignoredLines contains lines that should be excluded from lines
var ignoredLines = []string{"=== RUN", "--- FAIL"}

// Line represents JSON line from "go test" tool's lines
type Line struct {
	Time    string
	Action  string
	Package string
	Test    string
	Output  string
}

// Lines is set of lines
type Lines []Line

// Failed returns packages with failed tests
func (lns Lines) Failed() FailureGroup {
	f := make(FailureGroup)
	outputs := make([]int, 0, len(lns))

	// step 1: collect failed tests
	for i, l := range lns {
		if l.Test == "" {
			continue
		}

		switch l.Action {
		case actionOutput:
			if !lineIgnored(l.Output) {
				outputs = append(outputs, i)
			}
		case actionFail:
			if !f.hasPackage(l.Package) {
				f[l.Package] = make(Failures, 0)
			}

			f[l.Package][l.Test] = make([]string, 0)
		default:
			continue
		}
	}

	// step 2: collect output from failed tests
	for _, i := range outputs {
		l := lns[i]
		if !f.hasTest(l.Package, l.Test) {
			continue
		}

		// TODO: group table tests ("test/table_test")
		f.appendReportLine(l.Package, l.Test, strings.TrimSpace(l.Output))
	}

	return f
}

// SkippedPackages returns list of packages without unit tests
func (lns Lines) SkippedPackages() []string {
	pkgs := make([]string, 0, len(lns))
	for _, l := range lns {
		if l.Action != actionSkip {
			continue
		}

		pkgs = append(pkgs, l.Package)
	}

	return pkgs
}

func (lns *Lines) AppendData(data []byte) error {
	var l Line
	if err := json.Unmarshal(data, &l); err != nil {
		return err
	}

	*lns = append(*lns, l)
	return nil
}

func Parse(data []byte) (l Line, err error) {
	err = json.Unmarshal(data, &l)
	return l, err
}

func lineIgnored(data string) bool {
	if data == "" {
		return false
	}

	for _, s := range ignoredLines {
		if strings.Contains(data, s) {
			return true
		}
	}

	return false
}
