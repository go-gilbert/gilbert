package build

import (
	"github.com/go-gilbert/gilbert-sdk"
)

// NewAction creates a new build action instance
func NewAction(scope sdk.ScopeAccessor, params sdk.ActionParams) (sdk.ActionHandler, error) {
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
