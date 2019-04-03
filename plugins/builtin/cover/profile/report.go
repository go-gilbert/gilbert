package profile

import (
	"fmt"
	"github.com/axw/gocov"
	"github.com/axw/gocov/gocovutil"
	"strings"
)

// Coverage is coverage report with total and reached statements count
type Coverage struct {
	// Total is total statements count
	Total int

	// Reached is covered statements count
	Reached int
}

// Percentage gets coverage in percents
func (c *Coverage) Percentage() float64 {
	return float64(c.Reached*100) / float64(c.Total)
}

func (c *Coverage) add(cv Coverage) {
	c.Total += cv.Total
	c.Reached += cv.Reached
}

// Report is coverage report from GoCov profile
type Report struct {
	Coverage
	Packages map[string]*PackageReport
}

// CheckCoverage checks if report satisfies coverage requirements
func (r *Report) CheckCoverage(threshold float64) error {
	coverage := r.Percentage()
	if coverage < threshold {
		return fmt.Errorf("code coverage is below %.2f%% (got %.2f%%)", threshold, coverage)
	}

	return nil
}

// FormatFull returns detailed report
func (r *Report) FormatFull() string {
	b := strings.Builder{}
	for pkgName, pkg := range r.Packages {
		_, _ = fmt.Fprintf(&b, "Package '%s' - %.2f%%\n", pkgName, pkg.Percentage())
		for fnName, fn := range pkg.Functions {
			_, _ = fmt.Fprintf(&b, "  - %s: %.2f%%\n", fnName, fn.Percentage())
		}
	}

	return b.String()
}

// FormatSimple returns simplified report
func (r *Report) FormatSimple() string {
	b := strings.Builder{}
	for pkgName, pkg := range r.Packages {
		_, _ = fmt.Fprintf(&b, "  - %s: %.2f%%\n", pkgName, pkg.Percentage())
	}

	return b.String()
}

// PackageReport is package coverage report
type PackageReport struct {
	Coverage
	Functions map[string]*Coverage
}

// Create creates a new report from GoCov profile
func Create(pkgs gocovutil.Packages) (r Report) {
	r.Packages = make(map[string]*PackageReport, len(pkgs))
	for _, pkg := range pkgs {
		cov := pkgCoverage(pkg)
		r.add(cov.Coverage)
		r.Packages[pkg.Name] = cov
	}

	return r
}

func pkgCoverage(pkg *gocov.Package) *PackageReport {
	report := &PackageReport{}
	if len(pkg.Functions) == 0 {
		return report
	}

	fns := make(map[string]*Coverage, len(pkg.Functions))
	for _, fn := range pkg.Functions {
		c := Coverage{}
		c.Total, c.Reached = fnCoverage(fn)
		report.Coverage.add(c)
		fns[fn.Name] = &c
	}

	report.Functions = fns
	return report
}

func fnCoverage(fn *gocov.Function) (total, reached int) {
	total = len(fn.Statements)
	if total == 0 {
		return total, 0
	}

	reached = reachedStatements(fn.Statements)
	return total, reached
}

func reachedStatements(s []*gocov.Statement) (count int) {
	for _, st := range s {
		if st.Reached > 0 {
			count++
		}
	}

	return count
}
