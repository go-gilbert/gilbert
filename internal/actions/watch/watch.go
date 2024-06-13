package watch

import (
	"github.com/go-gilbert/gilbert-sdk"
)

// NewAction creates a new watch action handler instance
func NewAction(scope sdk.ScopeAccessor, rawParams sdk.ActionParams) (sdk.ActionHandler, error) {
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
