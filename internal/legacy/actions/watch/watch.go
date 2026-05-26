package watch

import (
	"github.com/go-gilbert/gilbert/internal/legacy/manifest"
	"github.com/go-gilbert/gilbert/internal/legacy/runner"
	"github.com/go-gilbert/gilbert/internal/legacy/scope"
)

// NewAction creates a new watch action handler instance
func NewAction(scope *scope.Scope, rawParams manifest.ActionParams) (runner.ActionHandler, error) {
	params, err := parseParams(rawParams, scope)
	if err != nil {
		return nil, err
	}

	return &Action{
		params: *params,
		scope:  scope,
		done:   make(chan bool),
	}, nil
}
