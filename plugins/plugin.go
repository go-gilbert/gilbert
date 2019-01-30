package plugins

import (
	"github.com/x1unix/guru/env"
	"github.com/x1unix/guru/logging"
	"github.com/x1unix/guru/manifest"
)

type Jar map[string]Plugin

type PluginFactory func(*env.Context, manifest.RawParams, logging.Logger) (Plugin, error)

type Plugin interface {
	Call() error
}
