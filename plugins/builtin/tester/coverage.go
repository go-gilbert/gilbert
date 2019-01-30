package tester

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type CoverResult struct {
	Package      string
	HasTestFiles bool
	Failed       bool
	Coverage     float64
}

var stringDelimiter string
var (
	re       *regexp.Regexp
	coverExp *regexp.Regexp
)

const (
	noTests     = "?"
	testPassed  = "ok"
	testFail    = "fail"
	noTestFiles = "[no test files]"
)

func init() {
	if runtime.GOOS == "windows" {
		stringDelimiter = "\r\n"
	} else {
		stringDelimiter = "\n"
	}

	re = regexp.MustCompile(`(?m)^(ok|\?|fail) *([0-9a-z/.]+)`)
	coverExp = regexp.MustCompile(`(?m)coverage: ([0-9.]+)\%`)
}

type CoverageChecker struct {
	params *CoverageParams
}

func (c *CoverageChecker) parseTestInfo(line string, r *CoverResult) error {
	groups := re.FindAllStringSubmatch(line, -1)
	if len(groups) == 0 {
		return fmt.Errorf("cannot parse data from go test output - '%v'", line)
	}

	matches := groups[0]
	if len(matches) < 3 {
		return fmt.Errorf("matches count mismatch")
	}

	switch matches[1] {
	case testPassed:
		r.HasTestFiles = true
	case testFail:
		r.HasTestFiles = true
		r.Failed = true
	default:
		r.HasTestFiles = false
	}

	r.Package = matches[2]
	return nil
}

func (c *CoverageChecker) extractCoverage(line string, r *CoverResult) (err error) {
	groups := re.FindAllStringSubmatch(line, -1)
	if len(groups) == 0 {
		return fmt.Errorf("cannot parse coverage from go test output - '%v'", line)
	}

	matches := groups[0]
	if len(matches) < 2 {
		return fmt.Errorf("matches count mismatch in coverage")
	}

	r.Coverage, err = strconv.ParseFloat(matches[1], 32)
	return err
}

func (c *CoverageChecker) parseLine(line string) (r *CoverResult, err error) {
	r = new(CoverResult)
	if err := c.parseTestInfo(line, r); err != nil {
		return nil, err
	}

	if err := c.extractCoverage(line, r); err != nil {
		return nil, fmt.Errorf("failed to parse coverage from string, %v", err)
	}

	return r, err
}

func (c *CoverageChecker) parseLines(rawOutput []byte) ([]*CoverResult, error) {
	lines := strings.Split(string(rawOutput), stringDelimiter)
	results := make([]*CoverResult, 0, len(lines))
	for i, line := range lines {
		result, err := c.parseLine(line)
		if err != nil {
			return nil, err
		}

		if c.params.IgnoreUncovered && !result.HasTestFiles {
			continue
		}

		results[i] = result
	}

	return results, nil
}

func (c *CoverageChecker) isIgnored(r *CoverResult) bool {
	for _, ignorePath := range c.params.Ignore {
		if strings.HasPrefix(r.Package, ignorePath) {
			return true
		}
	}

	return false
}

func (c *CoverageChecker) checkCoverage(output []byte) error {
	results, err := c.parseLines(output)
	if err != nil {
		return err
	}

	resultsCount := len(results)
	totalCoverage := 0.0
	for _, result := range results {
		totalCoverage += result.Coverage
	}

	gotCoverage := float32(totalCoverage) / float32(resultsCount)
	if gotCoverage < c.params.Threshold {
		return fmt.Errorf("total code coverage is below minimum threshold of %d% (got %d%)", c.params.Threshold, gotCoverage)
	}

	return nil
}
