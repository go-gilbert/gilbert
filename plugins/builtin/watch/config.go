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
	Job          *manifest.Job
}

func parseParams(raw manifest.RawParams, scope *scope.Scope) (*params, error) {
	p := params{
		DebounceTime: defaultDebounceTime,
	}
	if err := mapstructure.Decode(raw, &p); err != nil {
		return nil, fmt.Errorf("failed to read configuration: %s", err)
	}

	if p.Job == nil {
		return nil, fmt.Errorf("job to run is not defined")
	}

	if err := scope.Scan(&p.Path); err != nil {
		return nil, err
	}

	if p.Path == "" {
		return nil, fmt.Errorf("watch path is undefined, please set path to watch in 'path' parameter")
	}

	return &p, nil
}
