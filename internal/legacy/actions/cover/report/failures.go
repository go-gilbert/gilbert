package report

// Failures contains output lines of failed package tests
type Failures map[string][]string

// FailureGroup is a set of packages with failed tests
type FailureGroup map[string]Failures

func (f FailureGroup) hasPackage(pkg string) bool {
	_, ok := f[pkg]
	return ok
}

func (f FailureGroup) hasTest(pkg, test string) bool {
	if _, ok := f[pkg]; !ok {
		return false
	}

	if _, ok := f[pkg][test]; !ok {
		return false
	}

	return true
}

func (f FailureGroup) appendReportLine(pkg, test, line string) {
	f[pkg][test] = append(f[pkg][test], line)
}
