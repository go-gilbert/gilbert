package watch

import (
	"github.com/go-gilbert/gilbert-sdk"
)

// NewWatchPlugin creates a new plugin instance
func NewWatchPlugin(scope sdk.ScopeAccessor, rawParams sdk.PluginParams, log sdk.Logger) (sdk.Plugin, error) {
	params, err := parseParams(rawParams, scope)
	if err != nil {
		return nil, err
	}

	p, err := newPlugin(scope, *params, log)
	if err != nil {
		return nil, err
	}
	return p, nil
}
