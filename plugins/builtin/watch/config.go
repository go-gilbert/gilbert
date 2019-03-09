package watch

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
)

var defaultDebounceTime = manifest.Period(1000)

type params struct {
	Path         string
	DebounceTime manifest.Period
	Run          manifest.Job
}

func parseParams(raw manifest.RawParams, scope *scope.Scope) (*params, error) {
	p := params{
		DebounceTime: defaultDebounceTime,
	}
	if err := mapstructure.Decode(raw, &p); err != nil {
		return nil, fmt.Errorf("failed to read configuration: %s", err)
	}

	if err := scope.Scan(&p.Path); err != nil {
		return nil, err
	}

	return &p, nil
}
