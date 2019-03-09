package watch

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
	"os"
)

type params struct {
	Path string
	Run  manifest.Job
}

func (p *params) pathValid() error {
	_, err := os.Stat(p.Path)
	if os.IsNotExist(err) {
		return fmt.Errorf("object '%s' doesn't exists", err)
	}

	return err
}

func parseParams(raw manifest.RawParams, scope *scope.Scope) (*params, error) {
	p := params{}
	if err := mapstructure.Decode(raw, &p); err != nil {
		return nil, fmt.Errorf("failed to read configuration: %s", err)
	}

	if err := scope.Scan(&p.Path); err != nil {
		return nil, err
	}

	return &p, nil
}
