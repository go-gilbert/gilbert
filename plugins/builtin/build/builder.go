package build

import (
	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/guru/env"
	"github.com/x1unix/guru/logging"
	"github.com/x1unix/guru/manifest"
	"github.com/x1unix/guru/plugins"
)

func NewBuildPlugin(context *env.Context, params manifest.RawParams, log logging.Logger) (plugins.Plugin, error) {
	p := newParams()
	if err := mapstructure.Decode(params, &p); err != nil {
		return nil, manifest.NewPluginConfigError("build", err)
	}

	return &Plugin{
		context: context,
		params:  p,
		log:     log,
	}, nil
}
