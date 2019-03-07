package runner

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"

	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/plugins/builtin"
	"github.com/x1unix/gilbert/runner/job"
)

var errNoTaskHandler = fmt.Errorf("no task handler defined, please define task handler in 'plugin' or 'mixin' paramerer")

// taskRunner is TaskRunner implementation
type taskRunner struct {
	plugins          map[string]plugins.PluginFactory
	manifest         *manifest.Manifest
	CurrentDirectory string
	log              logging.Logger
	subLogger        logging.Logger
	context          context.Context
	cancelFn         context.CancelFunc
	wg               *sync.WaitGroup
}

// PluginByName gets plugin by name
func (t *taskRunner) PluginByName(pluginName string) (p plugins.PluginFactory, err error) {
	p, ok := t.plugins[pluginName]

	if !ok {
		err = fmt.Errorf("plugin '%s' not found", pluginName)
		return
	}

	return
}

// Stop stops task runner
func (t *taskRunner) Stop() {
	if t.cancelFn != nil {
		t.cancelFn()
	}
}

// RunTask execute task by name
func (t *taskRunner) RunTask(taskName string) (err error) {
	task, ok := t.manifest.Tasks[taskName]
	if !ok {
		return fmt.Errorf("task '%s' doesn't exists", taskName)
	}

	t.log.Log("Running task '%s'...", taskName)
	steps := len(task)

	t.context, t.cancelFn = context.WithCancel(context.Background())
	sl := t.subLogger.SubLogger()

	// Set waitgroup and buff channel for async jobs.
	wg := &sync.WaitGroup{}
	asyncJobsCount := task.AsyncJobsCount()
	asyncErrors := make(chan error, asyncJobsCount)
	t.subLogger.Log("%d async jobs in task", asyncJobsCount)

	defer func() {
		// Wait for unfinished async tasks
		// and collect results from async jobs
		wg.Wait()
		close(asyncErrors)
		for err := range asyncErrors {
			if err != nil {
				t.subLogger.Error("async job returned error: %s", err)
			}
		}
	}()

	for jobIndex, j := range task {
		currentStep := jobIndex + 1
		descr := j.FormatDescription()
		t.subLogger.Log("Step %d of %d: %s", currentStep, steps, descr)
		var err error
		ctx := job.NewRunContext(nil, sl, t.context)
		ctx.SetWaitGroup(wg)

		if j.Async {
			wg.Add(1)
			go t.runJob(&j, ctx)
		} else {
			err = t.startJobAndWait(&j, ctx)
		}
		if err != nil {
			return fmt.Errorf("task '%s' returned an error on step %d: %v", taskName, currentStep, err)
		}
	}

	return nil
}

func (t *taskRunner) startJobAsync(job *manifest.Job, ctx job.RunContext, errorHandler func(error)) {
	go t.runJob(job, ctx)
	select {
	case err := <-ctx.Error:
		errorHandler(err)
	}
}

func (t *taskRunner) startJobAndWait(job *manifest.Job, ctx job.RunContext) error {
	go t.runJob(job, ctx)
	select {
	case err := <-ctx.Error:
		return err
	}
}

// runJob execute specified job
func (t *taskRunner) runJob(j *manifest.Job, ctx job.RunContext) {
	s := scope.CreateScope(t.CurrentDirectory, j.Vars).
		AppendGlobals(t.manifest.Vars).
		AppendVariables(ctx.RootVars)

	// check if job should be run
	if !t.shouldRunJob(j, s) {
		ctx.Logger.Info("step was skipped")
		ctx.Success()
		return
	}

	// Wait if necessary
	if j.Delay > 0 {
		ctx.Logger.Debug("Job delay defined, waiting %dms...", j.Delay)
		time.Sleep(j.Delay.ToDuration())
	}

	if j.Deadline > 0 {
		// Add timeout if requested
		ctx = ctx.WithTimeout(j.Deadline.ToDuration())
	}

	execType := j.Type()
	switch execType {
	case manifest.ExecPlugin:
		t.applyJobPlugin(s, j, &ctx)
	case manifest.ExecMixin:
		t.execJobWithMixin(j, s, &ctx)
	default:
		ctx.Fail(errNoTaskHandler)
	}
}

func (t *taskRunner) applyJobPlugin(s *scope.Scope, j *manifest.Job, ctx *job.RunContext) {
	factory, err := t.PluginByName(j.PluginName)
	if err != nil {
		ctx.Fail(err)
		return
	}

	plugin, err := factory(s, j.Params, ctx.Logger)
	if err != nil {
		ctx.Fail(fmt.Errorf("failed to apply plugin '%s': %v", j.PluginName, err))
	}

	// Handle stop event
	// Event may arrive on SIGKILL or when timeout reached
	go func() {
		select {
		case <-ctx.Context.Done():
			ctx.Logger.Debug("%s: stop signal received", j.PluginName)
			ctx.Fail(plugin.Cancel())
		}
	}()

	// Call plugin
	if err := plugin.Call(ctx, t); err != nil {
		ctx.Fail(err)
	}
}

// execJobWithMixin constructs a task from job with mixin and runs it
//
// requires subLogger instance to create cascade logging output
func (t *taskRunner) execJobWithMixin(j *manifest.Job, s *scope.Scope, ctx *job.RunContext) {
	mx, ok := t.manifest.Mixins[j.MixinName]
	if !ok {
		ctx.Fail(fmt.Errorf("mixin '%s' doesn't exists", j.MixinName))
		return
	}

	// Create a task from mixin and job params
	ctx.Logger.Debug("create sub-task from mixin '%s'", j.MixinName)
	task := mx.ToTask(s.Variables)
	if err := t.runSubTask(task, s, ctx); err != nil {
		ctx.Fail(err)
		return
	}

	ctx.Success()
}

// runSubTask used to run sub-tasks created by parent job
//
// parentCtx used to expand task base properties (like description, etc.)
//
// subLogger used to create stack of log lines
func (t *taskRunner) runSubTask(task manifest.Task, parentScope *scope.Scope, parentCtx *job.RunContext) error {
	steps := len(task)

	// Set waitgroup and buff channel for async jobs.
	wg := &sync.WaitGroup{}
	asyncJobsCount := task.AsyncJobsCount()
	asyncErrors := make(chan error, asyncJobsCount)
	parentCtx.Logger.Log("%d async jobs in task", asyncJobsCount)

	defer func() {
		// Wait for unfinished async tasks
		// and collect results from async jobs
		wg.Wait()
		close(asyncErrors)
		for err := range asyncErrors {
			if err != nil {
				parentCtx.Logger.Error("async j returned error: %s", err)
			}
		}
	}()

	for jobIndex, j := range task {
		currentStep := jobIndex + 1

		// sub task label can contain template expressions (e.g. mixin step description)
		// so we should try to parse it
		descr := j.FormatDescription()
		if parsed, err := parentScope.ExpandVariables(descr); err != nil {
			parentCtx.Logger.Error("description parse error: %s", err)
		} else {
			descr = parsed
		}

		if steps > 1 {
			// show total steps count only if more than one step provided
			parentCtx.Logger.Info("- [%d/%d] %s", currentStep, steps, descr)
		} else {
			parentCtx.Logger.Info("- %s", descr)
		}

		ctx := parentCtx.ChildContext()
		if j.Async {
			wg.Add(1)
			go t.runJob(&j, ctx)
			continue
		}

		if err := t.startJobAndWait(&j, ctx); err != nil {
			return fmt.Errorf("%v (sub-task step %d)", err, currentStep)
		}
	}

	return nil
}

func (t *taskRunner) shouldRunJob(job *manifest.Job, ctx *scope.Scope) bool {
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

// NewTaskRunner creates a new TaskRunner instance
func NewTaskRunner(man *manifest.Manifest, cwd string, writer logging.Logger) TaskRunner {
	t := &taskRunner{
		plugins:          builtin.DefaultPlugins,
		manifest:         man,
		CurrentDirectory: cwd,
		log:              writer,
		subLogger:        writer.SubLogger(),
	}

	return t
}
