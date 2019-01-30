package runner

import (
	"fmt"
	"github.com/x1unix/guru/env"
	"github.com/x1unix/guru/manifest"

	"github.com/x1unix/guru/logging"
	"github.com/x1unix/guru/plugins"
	"github.com/x1unix/guru/plugins/builtin"
)

type TaskRunner struct {
	Plugins          map[string]plugins.PluginFactory
	Manifest         *manifest.Manifest
	CurrentDirectory string
	Log              logging.Logger
}

func (t *TaskRunner) PluginByName(pluginName string) (p plugins.PluginFactory, err error) {
	p, ok := t.Plugins[pluginName]

	if !ok {
		err = fmt.Errorf("plugin '%s' not found", pluginName)
		return
	}

	return
}

func (t *TaskRunner) RunJob(job *manifest.Job) error {
	ctx := env.CreateContext(t.CurrentDirectory, job.Vars).
		AppendGlobals(t.Manifest.Vars)

	if job.InvokesPlugin() {
		factory, err := t.PluginByName(job.Plugin)
		if err != nil {
			return err
		}

		plugin, err := factory(ctx, job.Params, t.Log.SubLogger())
		if err != nil {
			return fmt.Errorf("failed to apply plugin '%s': %v", job.Plugin, err)
		}
		return plugin.Call()
	}

	return fmt.Errorf("nested task invocation support is not supported, please use plugins for jobs")
}

func (t *TaskRunner) TaskByName(taskName string) (taskPtr *manifest.Task, ok bool) {
	task, ok := t.Manifest.Tasks[taskName]
	if !ok {
		return
	}

	taskPtr = &task
	return
}

func NewTaskRunner(man *manifest.Manifest, cwd string, writer logging.Logger) *TaskRunner {
	t := &TaskRunner{
		Plugins:          builtin.DefaultPlugins,
		Manifest:         man,
		CurrentDirectory: cwd,
		Log:              writer,
	}

	return t
}
