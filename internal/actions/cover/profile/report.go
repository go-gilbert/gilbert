package profile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/axw/gocov"
	"github.com/axw/gocov/gocovutil"
)

// Packages is a set of reports for each package
type Packages map[string]*PackageReport

// Names returns a slice of package names
func (p Packages) Names() []string {
	out := make([]string, 0, len(p))
	for k := range p {
		out = append(out, k)
	}

	return out
}

// Sort sorts packages by specified criteria
func (p Packages) Sort(by string, desc bool) []string {
	keys := p.Names()
	var sortFn sortSelector
	if by == ByCoverage {
		sortFn = pkgByPercentage(p)
	} else {
		sortFn = byName
	}

	s := &mapSorter{desc: desc, keys: keys, by: sortFn}
	sort.Sort(s)
	return keys
}

// Coverage is coverage report with total and reached statements count
type Coverage struct {
	// Total is total statements count
	Total int

	// Reached is covered statements count
	Reached int
}

// Percentage gets coverage in percents
func (c *Coverage) Percentage() float64 {
	if c.Reached == 0 {
		return 0
	}
	return float64(c.Reached*100) / float64(c.Total)
}

func (c *Coverage) add(cv Coverage) {
	c.Total += cv.Total
	c.Reached += cv.Reached
}

// Report is coverage report from GoCov profile
type Report struct {
	Coverage
	Packages Packages
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
func (r *Report) FormatFull(orderProp string, desc bool) string {
	b := strings.Builder{}
	pkgNames := r.Packages.Sort(orderProp, desc)
	for _, pkgName := range pkgNames {
		pkg := r.Packages[pkgName]
		_, _ = fmt.Fprintf(&b, "  Package '%s' - %.2f%%\n", pkgName, pkg.Percentage())

		fnNames := pkg.Sort(orderProp, desc)
		for _, fnName := range fnNames {
			fn := pkg.Functions[fnName]
			_, _ = fmt.Fprintf(&b, "    - %s: %.2f%%\n", fnName, fn.Percentage())
		}
	}

	return b.String()
}

// FormatSimple returns simplified report
func (r *Report) FormatSimple(orderProp string, desc bool) string {
	b := strings.Builder{}

	pkgNames := r.Packages.Sort(orderProp, desc)
	for _, pkgName := range pkgNames {
		pkg := r.Packages[pkgName]
		_, _ = fmt.Fprintf(&b, "  - %s: %.2f%%\n", pkgName, pkg.Percentage())
	}

	return b.String()
}

// PackageReport is package coverage report
type PackageReport struct {
	Coverage
	Functions map[string]*Coverage
}

func (p *PackageReport) names() []string {
	out := make([]string, 0, len(p.Functions))
	for k := range p.Functions {
		out = append(out, k)
	}

	return out
}

// Sort sorts package report data by specified criteria
func (p *PackageReport) Sort(by string, desc bool) []string {
	keys := p.names()
	var sortFn sortSelector
	if by == ByCoverage {
		sortFn = reportByPercentage(p)
	} else {
		sortFn = byName
	}

	s := &mapSorter{desc: desc, keys: keys, by: sortFn}
	sort.Sort(s)
	return keys
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
