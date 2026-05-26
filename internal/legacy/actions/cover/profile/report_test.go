package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/axw/gocov/gocovutil"
	"github.com/stretchr/testify/assert"
)

type packagesFile struct {
	Packages gocovutil.Packages
}

func getCoverProfile(t *testing.T) gocovutil.Packages {
	var out packagesFile
	data, err := os.ReadFile(filepath.Join(".", "testdata", "cover.json"))
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}

	return out.Packages
}

func TestReport_CheckCoverage(t *testing.T) {
	expPercentage := fmt.Sprintf("%.2f", 61.90)
	expected := Coverage{
		Total:   105,
		Reached: 65,
	}

	pkg := getCoverProfile(t)
	report := Create(pkg)
	assert.Equal(t, expected, report.Coverage)

	gotPercentage := fmt.Sprintf("%.2f", report.Coverage.Percentage())
	assert.Equal(t, expPercentage, gotPercentage)
}
