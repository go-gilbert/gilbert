package cover

import (
	"fmt"

	"github.com/go-gilbert/gilbert/actions/cover/profile"
)

// toolArgsPrefixSize is prefix args count for 'go tool cover' command
//
// go test -coverprofile=/tmp/cover -json ./services/foo ./services/bar./services/baz
const toolArgsPrefixSize = 3

type params struct {
	Threshold     float64   `mapstructure:"threshold"`
	Report        bool      `mapstructure:"reportCoverage"`
	ShowUncovered bool      `mapstructure:"showUncovered"`
	FullReport    bool      `mapstructure:"fullReport"`
	Packages      []string  `mapstructure:"packages"`
	Sort          sortParam `mapstructure:"sort"`
}

func (p *params) validate() error {
	if p.Threshold > 100 || p.Threshold < 0 {
		return fmt.Errorf("coverage threshold should be between 0 and 100 (got %f)", p.Threshold)
	}

	if p.Sort.By != profile.ByName && p.Sort.By != profile.ByCoverage {
		return fmt.Errorf("unsupported sort key '%s' (expected %s or %s)", p.Sort.By, profile.ByCoverage, profile.ByName)
	}

	return nil
}

type sortParam struct {
	By   string `mapstructure:"by"`
	Desc bool   `mapstructure:"desc"`
}

func newParams() params {
	return params{
		Threshold:     0.0,
		Report:        false,
		ShowUncovered: false,
		Sort: sortParam{
			By:   profile.ByCoverage,
			Desc: true,
		},
	}
}
