package runner

import "github.com/x1unix/gilbert/plugins"

// TaskRunner runs tasks from manifest file
type TaskRunner interface {
	PluginByName(pluginName string) (p plugins.PluginFactory, err error)
	RunTask(taskName string) (err error)
	Stop()
}
