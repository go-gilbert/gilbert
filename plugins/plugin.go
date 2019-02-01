package plugins

import (
	"github.com/x1unix/guru/logging"
	"github.com/x1unix/guru/manifest"
	"github.com/x1unix/guru/scope"
)

type Jar map[string]Plugin

type PluginFactory func(*scope.Context, manifest.RawParams, logging.Logger) (Plugin, error)

type Plugin interface {
	Call() error
}
