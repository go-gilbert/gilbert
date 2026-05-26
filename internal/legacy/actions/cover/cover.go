package cover

import (
	"fmt"
	"os"

	"github.com/go-gilbert/gilbert/internal/legacy/manifest"
	"github.com/go-gilbert/gilbert/internal/legacy/runner"
	"github.com/go-gilbert/gilbert/internal/legacy/scope"
)

const coverFilePattern = "gbcover*.out"

// NewAction creates a new cover action handler instance
func NewAction(scope *scope.Scope, params manifest.ActionParams) (runner.ActionHandler, error) {
	p := newParams()
	if err := params.Unmarshal(&p); err != nil {
		return nil, err
	}

	if err := p.validate(); err != nil {
		return nil, err
	}

	f, err := os.CreateTemp(os.TempDir(), coverFilePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create coverage temporary file: %s", err)
	}

	return &Action{
		scope:     scope,
		params:    p,
		alive:     true,
		coverFile: f,
	}, nil
}
