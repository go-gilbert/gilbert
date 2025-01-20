package report

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fixturePath = filepath.Join(".", "testdata", "report.json")

func TestLines_Failed(t *testing.T) {
	expected := FailureGroup{
		"github.com/go-gilbert/gilbert/actions/cover": Failures{
			"": []string{}, "TestParamsValidate": []string{},
			"TestParamsValidate/should_validate_threshold_above_100": []string{
				`params_test.go:48: error 'coverage threshold should be between 0 and 100 (got 101.000000)' should contain 'coverage threshold should be between 0 and 1001'`,
			},
		},
	}
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatal(err)
	}

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)

	lines := make(Lines, 0)
	for s.Scan() {
		data := s.Bytes()
		if err := lines.AppendData(data); err != nil {
			t.Error(err)
		}
	}

	if err := s.Err(); err != nil {
		t.Fatal(err)
	}

	failed := lines.Failed()
	assert.Equal(t, expected, failed)
}

func TestLines_SkippedPackages(t *testing.T) {
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatal(err)
	}

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)

	lines := make(Lines, 0)
	for s.Scan() {
		data := s.Bytes()
		if err := lines.AppendData(data); err != nil {
			t.Error(err)
		}
	}

	if err := s.Err(); err != nil {
		t.Fatal(err)
	}

	skipped := lines.SkippedPackages()
	assert.NotEmpty(t, skipped)
}
