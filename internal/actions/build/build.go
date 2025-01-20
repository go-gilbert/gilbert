package build

import (
	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/runner"
	"github.com/go-gilbert/gilbert/internal/scope"
)

// NewAction creates a new build action instance
func NewAction(scope *scope.Scope, params manifest.ActionParams) (runner.ActionHandler, error) {
	p := newParams()
	if err := params.Unmarshal(&p); err != nil {
		return nil, err
	}

	if err := scope.Scan(&p.Target.Os, &p.Target.Arch); err != nil {
		return nil, err
	}

	return &Action{
		scope:  scope,
		params: p,
	}, nil
}
