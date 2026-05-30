// Package engine implements task run engine and its core functionality.
package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type CommandProcessorFactory = func(*scope.Scope) expr.CommandProcessor

type Config struct {
	Logger                *log.Logger
	JobFile               *manifest.JobFile
	RootScope             *scope.Scope
	Shell                 *Shell
	CmdProcessorFactory   CommandProcessorFactory
	ActionHandlerProvider ActionHandlerProvider
	MaxConcurrentJobs     int
}

type Runner struct {
	logger         *log.Logger
	jobFile        *manifest.JobFile
	rootScope      *scope.Scope
	shell          *Shell
	cmdProcBuilder CommandProcessorFactory
	actionHandlers ActionHandlerProvider
}

func NewRunner(cfg Config) *Runner {
	return &Runner{
		logger:         cfg.Logger,
		jobFile:        cfg.JobFile,
		rootScope:      cfg.RootScope,
		cmdProcBuilder: cfg.CmdProcessorFactory,
		shell:          cfg.Shell,
		actionHandlers: cfg.ActionHandlerProvider,
	}
}

type forkContext struct {
	name fmt.Stringer
	kind manifest.JobKind
}

// forkScope creates a new scope child with its local env vars and working directory.
func (r *Runner) forkScope(ctx context.Context, parent *scope.Scope, params manifest.CommonRunParams, fk forkContext) (*scope.Scope, error) {
	s := parent.Fork().WithWorkDir(params.WorkDir)
	diags := manifest.AppendLazyEnvVarsToScope(ctx, s, params.Env)
	r.shell.Reporter.PrintDiagnostics(diags)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to expand environment variables for %s %q", fk.kind, fk.name)
	}

	return s, nil
}

type stringer string

func (s stringer) String() string {
	return string(s)
}

// RunTaskWithScope starts a runner with a given task and scope.
//
// Note: passed scope should have the same root as used by runner.
func (r *Runner) RunTaskWithScope(ctx context.Context, name string, s *scope.Scope) error {
	if s.Root != r.rootScope {
		return errors.New("passed scope has a different root")
	}

	t, ok := r.jobFile.Tasks[name]
	if !ok {
		return fmt.Errorf("task %q not found", name)
	}

	r.shell.Reporter.OnTaskStart(TaskStartEvent{TaskName: name})

	// initialize task scope with local env vars and work dir
	taskScope, err := r.forkScope(ctx, r.rootScope, t.CommonRunParams, forkContext{
		name: stringer(name),
		kind: manifest.JobKindTask,
	})
	if err != nil {
		return err
	}

	return r.runJobGroup(ctx, t.Jobs, taskScope)
}

func (r *Runner) runJobGroup(ctx context.Context, jobs []manifest.Job, s *scope.Scope) error {
	// vars for async jobs
	taskCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	var lastError error
	g := newAsyncJobGroup(r, cancelFn)
	for _, j := range jobs {
		if taskCtx.Err() != nil {
			break
		}

		if j.Async {
			g.schedule(taskCtx, j, s)
			continue
		}

		err := r.runJob(taskCtx, j, s)
		if err != nil {
			// TODO: impl allow failure
			// Remember failed error, terminate job context
			cancelFn()
			lastError = err
			break
		}
	}

	if !g.isInitialized() {
		return lastError
	}

	// Wait for async jobs
	if c := g.remainingCount.Load(); c > 0 {
		if lastError != nil {
			// Log failed sync tasks immediately
			r.logger.Error(lastError)
		}

		r.logger.Infof("waiting for %d async jobs to finish", c)
	}

	asyncErr := g.wait()
	if lastError != nil {
		// Sync job errors are primary.
		// Async errors are already logged.
		return lastError
	}

	return asyncErr
}

func (r *Runner) runJob(ctx context.Context, j manifest.Job, taskScope *scope.Scope) error {
	ep := expr.EvalParams{
		CommandProcessor: r.cmdProcBuilder(taskScope),
		Env:              taskScope,
	}

	ok, diag := shouldRunJob(ctx, ep, &j)
	if diag != nil {
		r.shell.Reporter.PrintDiagnostics(parsetypes.Diagnostics{diag})
		return fmt.Errorf("cannot check job condition: %w", diag.Err)
	}

	if !ok {
		r.logger.Debugw("job skipped", log.NewField("job", j.Handler))
		return nil
	}

	if len(j.Strategy.Matrix) != 0 {
		return r.runJobMatrix(ctx, j, taskScope)
	}

	jobScope, err := r.forkScope(ctx, taskScope, j.CommonRunParams, forkContext{
		name: j.Handler,
		kind: j.Kind,
	})
	if err != nil {
		return err
	}

	return r.handleJob(ctx, j, jobScope)
}

func (r *Runner) runJobMatrix(ctx context.Context, j manifest.Job, taskScope *scope.Scope) error {
	// Resolve matrix values
	ep := expr.EvalParams{
		CommandProcessor: r.cmdProcBuilder(taskScope),
		Env:              taskScope,
	}

	matParams, diags := resolveMatrixValues(ctx, ep, j.Strategy.Matrix)
	r.shell.Reporter.PrintDiagnostics(diags)
	if diags.HasError() {
		return fmt.Errorf("failed to resolve matrix values for step %q", j.Handler)
	}

	diags = validateMatrixExcludeRules(&j.Strategy, matParams)
	r.shell.Reporter.PrintDiagnostics(diags)
	if diags.HasError() {
		return fmt.Errorf(`invalid rule in "exclude" section`)
	}

	g, matCtx := errgroup.WithContext(ctx)
	if j.Strategy.MaxParallel > 0 {
		g.SetLimit(j.Strategy.MaxParallel)
	}

	// do Cartesian product and run each job
	for labels, values := range innerJoinMatrix(matParams) {
		if testMatrixExcluded(&j.Strategy, values) {
			r.logger.Debugw(
				"matrix config skipped",
				log.NewField("job", j.Handler),
				log.NewField("keys", labels),
				log.NewField("vals", values),
			)
			continue
		}

		// env vars may refer to matrix values, thus use separate scope per matrix job.
		jobScope := taskScope.Fork().WithMatrixValues(labels, values)
		jobScope, err := r.forkScope(ctx, jobScope, j.CommonRunParams, forkContext{
			name: j.Handler,
			kind: j.Kind,
		})
		if err != nil {
			return err
		}

		// TODO: support fail-fast?
		g.Go(func() error {
			if matCtx.Err() != nil {
				return nil
			}

			err := r.handleJob(matCtx, j, jobScope)
			if err == nil {
				return nil
			}

			err = fmt.Errorf("job %q (%s) returned an error: %w", j.Handler, formatMatParams(jobScope.MatrixValues), err)
			if !j.ContinueOnError {
				return err
			}

			// Don't log context canceled error
			if !errors.Is(err, context.Canceled) {
				r.logger.Error(err)
			}

			return nil
		})
	}

	return g.Wait()
}

// runSubtask starts mixin or task as a sub-task with its own separate scope.
func (r *Runner) runSubtask(ctx context.Context, j manifest.Job, jobScope *scope.Scope, kind manifest.JobKind) error {
	name := j.Handler.Name

	var (
		group *manifest.JobGroup
		ok    bool
	)
	switch kind {
	case manifest.JobKindMixin:
		group, ok = r.jobFile.Mixins[name]
	case manifest.JobKindTask:
		group, ok = r.jobFile.Tasks[name]
	default:
		panic("internal error: runSubtask: bad subtask kind: " + kind.String())
	}

	if !ok {
		return fmt.Errorf("%s %q doesn't exist", kind, name)
	}

	inputs, diags := manifest.MapArgsToInputs(ctx, manifest.MapArgsParams{
		Values:  j.Args,
		Spec:    group.Inputs,
		EnvVars: jobScope.Globals.Env,
		EvalParams: expr.EvalParams{
			CommandProcessor: r.cmdProcBuilder(jobScope),
			Env:              jobScope,
		},
	})
	r.shell.Reporter.PrintDiagnostics(diags)
	if diags.HasError() {
		return fmt.Errorf("invalid input parameters for %s %q", kind, name)
	}

	if kind == manifest.JobKindTask {
		r.shell.Reporter.OnTaskStart(TaskStartEvent{
			TaskName:  name,
			IsSubTask: true,
		})
	}

	// job working directory takes precedense.
	// use task/mixin workdir if runJob didn't change it before
	s := jobScope.Fork()
	if j.WorkDir == "" {
		s = s.WithWorkDir(group.WorkDir)
	}

	// Remove locals and prepare inputs.
	s.Inputs = inputs
	s.MatrixValues = nil
	s.EventData = nil

	// append task/mixin env vars
	diags = manifest.AppendLazyEnvVarsToScope(ctx, s, group.Env)
	r.shell.Reporter.PrintDiagnostics(diags)
	if diags.HasError() {
		return fmt.Errorf("failed to expand environment variables for %s %q", kind, group.Name)
	}

	err := r.runJobGroup(ctx, group.Jobs, s)
	if err != nil {
		return fmt.Errorf("%s %q returned error: %w", kind, name, err)
	}

	return nil
}

func (r *Runner) handleJob(ctx context.Context, j manifest.Job, jobScope *scope.Scope) error {
	if j.Kind != manifest.JobKindAction {
		return r.runSubtask(ctx, j, jobScope, j.Kind)
	}

	jobName := j.Handler.String()
	hResult := r.actionHandlers.GetActionHandler(ctx, j.Handler, ActionParams{
		Logger: r.logger.Named(jobName),
		Shell:  r.shell,
		Scope:  jobScope,
		Args:   j.Args,
		EvalParams: expr.EvalParams{
			CommandProcessor: r.cmdProcBuilder(jobScope),
			Env:              jobScope,
		},
		Outputs: []string{},
	})
	r.shell.Reporter.PrintDiagnostics(hResult.Diagnostics)
	if hResult.Error != nil {
		return hResult.Error
	}

	if j.Delay != 0 {
		r.logger.Debugw(
			"wait for job delay to finish",
			log.NewField("job", j.Handler),
			log.NewField("delay", j.Delay),
		)
	}
	if err := j.WaitForDelay(ctx); err != nil {
		return err
	}

	r.shell.Reporter.OnJobStart(JobStartEvent{
		JobName:          jobName,
		MatrixParameters: jobScope.MatrixValues,
	})

	runCtx, cancelFn := j.WrapContext(ctx)
	defer cancelFn()

	emitter := r.newSignalEmitter(jobName, jobScope, j.Hooks)
	err := hResult.Handler.HandleAction(runCtx, emitter)
	emitter.callErrorHook(ctx, err)
	return err
}

// signalEmitterPrivate is a wrapper interface around public's SignalEmiiter.
//
// Contains private functionality restricted to runner internal use.
type signalEmitterPrivate interface {
	SignalEmitter

	callErrorHook(ctx context.Context, err error)
}

func (r *Runner) newSignalEmitter(sender string, parentScope *scope.Scope, hooks manifest.SignalHooks) signalEmitterPrivate {
	if len(hooks) == 0 {
		return noopSignalEmitter{}
	}

	return &signalEmitter{
		sender:      sender,
		runner:      r,
		hooks:       hooks,
		parentScope: parentScope,
	}
}

type signalEmitter struct {
	sender      string
	runner      *Runner
	parentScope *scope.Scope
	hooks       manifest.SignalHooks
}

const errSignalTimeout = 5 * time.Second

func (em *signalEmitter) callErrorHook(ctx context.Context, err error) {
	if err == nil {
		return
	}

	// early check to avoid allocating arg map if not needed.
	_, ok := em.hooks["error"]
	if !ok {
		em.runner.logger.Debug("no error handler hooks, skip")
		return
	}

	if ctx.Err() != nil {
		// if canceled - replace with deadline to allow cleanup hooks to run
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), errSignalTimeout)
		defer cancel()
	}

	em.runHooks(ctx, "error", map[string]any{
		"error": err.Error(),
	})
}

func (em *signalEmitter) EmitSignal(ctx context.Context, name string, data map[string]any) error {
	if err := validateSignalIsAllowed(name); err != nil {
		return err
	}

	em.runHooks(ctx, name, data)
	return nil
}

func (em *signalEmitter) runHooks(ctx context.Context, name string, data map[string]any) {
	jobs, ok := em.hooks[name]
	if !ok {
		em.runner.logger.Debugf("no hooks for signal %q, skip", name)
		return
	}

	s := em.parentScope.Fork()
	s.EventData = data

	em.runner.shell.Reporter.OnSignal(SignalEvent{
		SignalName: name,
		JobName:    em.sender,
		Args:       data,
	})

	err := em.runner.runJobGroup(ctx, jobs, s)
	if err != nil {
		// signal handlers are allowed to fail
		em.runner.logger.Warnf("signal %q returned an error: %s", name, err)
	}
}

type noopSignalEmitter struct{}

func (_ noopSignalEmitter) callErrorHook(ctx context.Context, err error) {
	// NOOP
}

func (_ noopSignalEmitter) EmitSignal(ctx context.Context, name string, data map[string]any) error {
	return validateSignalIsAllowed(name)
}

func validateSignalIsAllowed(name string) error {
	if name == "error" {
		return fmt.Errorf("sending of internal signal is not allowed: %q", name)
	}

	return nil
}
