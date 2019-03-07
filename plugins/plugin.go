package plugins

import (
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/runner"
	"github.com/x1unix/gilbert/runner/job"
	"github.com/x1unix/gilbert/scope"
)

// Jar is a collection of plugins
type Jar map[string]Plugin

// PluginFactory is plugin constructor
type PluginFactory func(*scope.Scope, manifest.RawParams, logging.Logger) (Plugin, error)

// Plugin represents Gilbert's plugin
type Plugin interface {
	// Call calls a plugin
	Call(ctx *job.RunContext, r runner.TaskRunner) error
}
