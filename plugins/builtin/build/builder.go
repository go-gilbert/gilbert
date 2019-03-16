package build

import (
	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/scope"
)

// NewBuildPlugin creates a new build plugin instance
func NewBuildPlugin(scope *scope.Scope, params manifest.RawParams, log log.Logger) (plugins.Plugin, error) {
	p := newParams()
	if err := mapstructure.Decode(params, &p); err != nil {
		return nil, manifest.NewPluginConfigError("build", err)
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
