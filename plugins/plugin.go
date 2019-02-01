package plugins

import (
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
)

type Jar map[string]Plugin

type PluginFactory func(*scope.Context, manifest.RawParams, logging.Logger) (Plugin, error)

type Plugin interface {
	Call() error
}
