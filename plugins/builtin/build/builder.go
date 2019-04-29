package build

import (
	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/manifest"
)

// NewBuildPlugin creates a new build plugin instance
func NewBuildPlugin(scope sdk.ScopeAccessor, params sdk.PluginParams, log sdk.Logger) (sdk.Plugin, error) {
	p := newParams()
	if err := params.Unmarshal(&p); err != nil {
		return nil, err
	}

	if err := scope.Scan(&p.Target.Os, &p.Target.Arch); err != nil {
		return nil, manifest.NewPluginConfigError("build", err)
	}

	return &Plugin{
		scope:  scope,
		params: p,
		log:    log,
	}, nil
}
