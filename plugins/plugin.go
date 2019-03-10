package plugins

import (
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
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
	Call(*job.RunContext, JobRunner) error

	// Cancel stops plugin execution
	Cancel(*job.RunContext) error
}

// JobRunner runs jobs
type JobRunner interface {
	// PluginByName returns plugin constructor
	PluginByName(pluginName string) (p PluginFactory, err error)

	// RunJob starts job in separate goroutine.
	//
	// Use ctx.Error channel to track job result and ctx.Cancel() to cancel it.
	RunJob(j manifest.Job, ctx *job.RunContext)
}
