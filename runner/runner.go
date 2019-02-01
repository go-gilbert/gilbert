package runner

import (
	"fmt"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"
	"strings"

	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/plugins/builtin"
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
	ctx := scope.CreateContext(t.CurrentDirectory, job.Vars).
		AppendGlobals(t.Manifest.Vars)

	// check if job should be run
	if !t.shouldRunJob(job, ctx) {
		t.Log.SubLogger().Info("Step was skipped")
		return nil
	}

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

func (t *TaskRunner) shouldRunJob(job *manifest.Job, ctx *scope.Context) bool {
	condCmd := strings.TrimSpace(job.Condition)
	if condCmd == "" {
		return true
	}

	l := t.Log.SubLogger()
	condCmd, err := ctx.ExpandVariables(condCmd)
	if err != nil {
		l.Error(err.Error())
		l.Warn("Failed to parse value inside 'if' expression, job will be skipped")
		return false
	}
	cmd := shell.PrepareCommand(condCmd)

	l.Debug("assert command: '%s'", condCmd)

	// Return false if command failed to start or returned bad exit code
	if err := cmd.Start(); err != nil {
		return false
	}

	if err := cmd.Wait(); err != nil {
		return false
	}

	return true
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
