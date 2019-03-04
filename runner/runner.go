package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"

	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/plugins/builtin"
)

// TaskRunner is task runner
type TaskRunner struct {
	Plugins          map[string]plugins.PluginFactory
	Manifest         *manifest.Manifest
	CurrentDirectory string
	log              logging.Logger
	subLogger        logging.Logger
}

// PluginByName gets plugin by name
func (t *TaskRunner) PluginByName(pluginName string) (p plugins.PluginFactory, err error) {
	p, ok := t.Plugins[pluginName]

	if !ok {
		err = fmt.Errorf("plugin '%s' not found", pluginName)
		return
	}

	return
}

// RunTask execute task by name
func (t *TaskRunner) RunTask(taskName string) error {
	task, ok := t.TaskByName(taskName)
	if !ok {
		return fmt.Errorf("task '%s' doesn't exists", taskName)
	}

	t.log.Log("Running task '%s'...", taskName)
	steps := len(*task)

	for jobIndex, job := range *task {
		currentStep := jobIndex + 1
		descr := ""
		if job.HasDescription() {
			descr = ": " + job.Description
		}

		logging.Log.SubLogger().Log("Step %d of %d%s", currentStep, steps, descr)
		err := t.runJob(&job, nil)
		if err != nil {
			return fmt.Errorf("task '%s' returned an error on step %d: %v", taskName, currentStep, err)
		}
	}

	return nil
}

// runJob execute specified job
//
// if ctx is nil, default context will be created
func (t *TaskRunner) runJob(job *manifest.Job, ctx *scope.Context) error {
	if ctx == nil {
		ctx = scope.CreateContext(t.CurrentDirectory, job.Vars).
			AppendGlobals(t.Manifest.Vars)
	}

	// check if job should be run
	if !t.shouldRunJob(job, ctx) {
		t.subLogger.SubLogger().Info("Step was skipped")
		return nil
	}

	// Wait if necessary
	if job.Delay > 0 {
		t.subLogger.SubLogger().Debug("Job delay defined, waiting %dms...", job.Delay)
		time.Sleep(time.Duration(job.Delay) * time.Millisecond)
	}

	if job.PluginName != nil {
		factory, err := t.PluginByName(*job.PluginName)
		if err != nil {
			return err
		}

		plugin, err := factory(ctx, job.Params, t.subLogger.SubLogger())
		if err != nil {
			return fmt.Errorf("failed to apply plugin '%s': %v", *job.PluginName, err)
		}
		return plugin.Call()
	}

	return fmt.Errorf("no task handler defined, please define task handler in 'plugin' parameter")
}

func (t *TaskRunner) shouldRunJob(job *manifest.Job, ctx *scope.Context) bool {
	condCmd := strings.TrimSpace(job.Condition)
	if condCmd == "" {
		return true
	}

	l := t.subLogger.SubLogger()
	condCmd, err := ctx.ExpandVariables(condCmd)
	if err != nil {
		l.Error(err.Error())
		l.Warn("Failed to parse value inside 'if' expression, job will be skipped")
		return false
	}
	cmd := shell.PrepareCommand(condCmd)

	l.Debug("Assert command: '%s'", condCmd)

	// Return false if command failed to start or returned bad exit code
	if err := cmd.Start(); err != nil {
		return false
	}

	if err := cmd.Wait(); err != nil {
		return false
	}

	return true
}

// TaskByName returns a task by name
func (t *TaskRunner) TaskByName(taskName string) (taskPtr *manifest.Task, ok bool) {
	task, ok := t.Manifest.Tasks[taskName]
	if !ok {
		return
	}

	taskPtr = &task
	return
}

// NewTaskRunner creates a new TaskRunner instance
func NewTaskRunner(man *manifest.Manifest, cwd string, writer logging.Logger) *TaskRunner {
	t := &TaskRunner{
		Plugins:          builtin.DefaultPlugins,
		Manifest:         man,
		CurrentDirectory: cwd,
		log:              writer,
		subLogger:        writer.SubLogger(),
	}

	return t
}
