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

// TaskRunner runs tasks
type TaskRunner struct {
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
func (t *TaskRunner) PluginByName(pluginName string) (p plugins.PluginFactory, err error) {
	p, ok := t.plugins[pluginName]

	if !ok {
		err = fmt.Errorf("plugin '%s' not found", pluginName)
		return
	}

	return
}

// Stop stops task runner
func (t *TaskRunner) Stop() {
	if t.cancelFn != nil {
		t.cancelFn()
	}
}

// RunTask execute task by name
func (t *TaskRunner) RunTask(taskName string) (err error) {
	task, ok := t.manifest.Tasks[taskName]
	if !ok {
		return fmt.Errorf("task '%s' doesn't exists", taskName)
	}

	t.log.Log("Running task '%s'...", taskName)
	steps := len(task)

	t.context, t.cancelFn = context.WithCancel(context.Background())
	sl := t.subLogger.SubLogger()

	// Set waitgroup and buff channel for async jobs.
	var tracker *asyncJobTracker
	asyncJobsCount := task.AsyncJobsCount()
	if asyncJobsCount > 0 {
		t.subLogger.Debug("%d async jobs in task", asyncJobsCount)
		tracker = newAsyncJobTracker(t, t.context, asyncJobsCount)
		go tracker.trackAsyncJobs()

		defer func() {
			// Wait for unfinished async tasks
			// and collect results from async jobs
			t.subLogger.Log("Waiting for %d async job(s) to complete", asyncJobsCount)
			if asyncErr := tracker.wait(); asyncErr != nil {
				if err == nil {
					// Report error only if no previous errors.
					// P.S - it's okay since all async errors were logged previously
					err = fmt.Errorf("task '%s' returned error in async job: %s", taskName, asyncErr)
				}
			}
		}()
	}

	for jobIndex, j := range task {
		currentStep := jobIndex + 1
		descr := j.FormatDescription()
		t.subLogger.Log("Step %d of %d: %s", currentStep, steps, descr)
		var err error
		ctx := job.NewRunContext(nil, sl, t.context)

		if j.Async {
			tracker.decorateJobContext(&ctx)
			go t.handleJob(j, ctx)
			continue
		}

		if err = t.startJobAndWait(j, ctx); err != nil {
			return fmt.Errorf("task '%s' returned an error on step %d: %v", taskName, currentStep, err)
		}
	}

	return err
}

// RunJob starts job in separate goroutine.
//
// Use ctx.Error channel to track job result and ctx.Cancel() to cancel it.
func (t *TaskRunner) RunJob(j manifest.Job, ctx job.RunContext) {
	go t.handleJob(j, ctx)
}

func (t *TaskRunner) startJobAndWait(job manifest.Job, ctx job.RunContext) error {
	go t.handleJob(job, ctx)
	// All child jobs (except async jobs) inherit parent job channel,
	// so we should close channel only if parent job was finished.
	if !ctx.IsChild() {
		defer close(ctx.Error)
	}

	err, ok := <-ctx.Error
	if !ok {
		ctx.Logger.Debug("Error: failed to read data from result channel")
		return nil
	}

	return err
}

// handleJob handles specified job
func (t *TaskRunner) handleJob(j manifest.Job, ctx job.RunContext) {
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

	// FIXME: exec.Cmd-based jobs are still alive even if timeout elapsed
	// occurs only with context.WithTimeout()
	if j.Deadline > 0 {
		ctx.Logger.Warn("Warning: 'deadline' is unstable")

		// Add timeout if requested
		ttl := j.Deadline.ToDuration()
		ctx = ctx.WithTimeout(ttl)

		// Add deadline listener
		go func() {
			select {
			case <-time.After(1 * time.Second):
				ctx.Logger.Warn("job deadline finish")
				ctx.Cancel()
			}
		}()
	}

	execType := j.Type()
	switch execType {
	case manifest.ExecPlugin:
		t.applyJobPlugin(s, j, &ctx)
	case manifest.ExecMixin:
		t.execJobWithMixin(j, s, &ctx)
	default:
		ctx.Result(errNoTaskHandler)
	}
}

func (t *TaskRunner) applyJobPlugin(s *scope.Scope, j manifest.Job, ctx *job.RunContext) {
	factory, err := t.PluginByName(j.PluginName)
	if err != nil {
		ctx.Result(err)
		return
	}

	plugin, err := factory(s, j.Params, ctx.Logger)
	if err != nil {
		ctx.Result(fmt.Errorf("failed to apply plugin '%s': %v", j.PluginName, err))
		return
	}

	// Handle stop event
	// Event may arrive on SIGKILL or when timeout reached
	go func() {
		select {
		case <-ctx.Context.Done():
			ctx.Logger.Debug("sent stop signal to '%s' plugin", j.PluginName)
			ctx.Result(plugin.Cancel(ctx))
		}
	}()

	// Call plugin and send result
	err = plugin.Call(ctx, t)
	ctx.Result(err)
}

// execJobWithMixin constructs a task from job with mixin and runs it
//
// requires subLogger instance to create cascade logging output
func (t *TaskRunner) execJobWithMixin(j manifest.Job, s *scope.Scope, ctx *job.RunContext) {
	mx, ok := t.manifest.Mixins[j.MixinName]
	if !ok {
		ctx.Result(fmt.Errorf("mixin '%s' doesn't exists", j.MixinName))
		return
	}

	// Create a task from mixin and job params
	ctx.Logger.Debug("create sub-task from mixin '%s'", j.MixinName)
	task := mx.ToTask(s.Variables)
	if err := t.runSubTask(task, s, ctx); err != nil {
		ctx.Result(err)
		return
	}

	ctx.Success()
}

// runSubTask used to run sub-tasks created by parent job
//
// parentCtx used to expand task base properties (like description, etc.)
//
// subLogger used to create stack of log lines
func (t *TaskRunner) runSubTask(task manifest.Task, parentScope *scope.Scope, parentCtx *job.RunContext) (err error) {
	steps := len(task)

	// Set waitgroup and buff channel for async jobs.
	var tracker *asyncJobTracker
	asyncJobsCount := task.AsyncJobsCount()
	if asyncJobsCount > 0 {
		parentCtx.Logger.Debug("%d async jobs in sub-task", asyncJobsCount)
		tracker = newAsyncJobTracker(t, parentCtx.Context, asyncJobsCount)
		go tracker.trackAsyncJobs()

		defer func() {
			// Wait for unfinished async tasks
			// and collect results from async jobs
			t.subLogger.Log("Waiting for %d async job(s) to complete", asyncJobsCount)
			if asyncErr := tracker.wait(); asyncErr != nil {

				if err == nil {
					// Report error only if no previous errors.
					// P.S - it's okay since all async errors were logged previously
					err = fmt.Errorf("async job returned error - %s", asyncErr)
				}
			}
		}()
	}

	for jobIndex, j := range task {
		currentStep := jobIndex + 1

		// sub task label can contain template expressions (e.g. mixin step description)
		// so we should try to parse it
		descr := j.FormatDescription()
		if parsed, perr := parentScope.ExpandVariables(descr); perr != nil {
			parentCtx.Logger.Error("description parse error: %s", perr)
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
			tracker.decorateJobContext(&ctx)
			go t.handleJob(j, ctx)
			continue
		}

		if err = t.startJobAndWait(j, ctx); err != nil {
			return fmt.Errorf("%v (sub-task step %d)", err, currentStep)
		}
	}

	return err
}

func (t *TaskRunner) shouldRunJob(job manifest.Job, ctx *scope.Scope) bool {
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

// NewTaskRunner creates a new task runner instance
func NewTaskRunner(man *manifest.Manifest, cwd string, writer logging.Logger) *TaskRunner {
	t := &TaskRunner{
		plugins:          builtin.DefaultPlugins,
		manifest:         man,
		CurrentDirectory: cwd,
		log:              writer,
		subLogger:        writer.SubLogger(),
	}

	return t
}
