// Package engine implements task run engine and its core functionality.
package engine

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
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

	r.shell.Reporter.OnTaskStart(name)
	return r.runJobGroup(ctx, t.Jobs, s)
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

	return r.handleJob(ctx, j, taskScope)
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

		// TODO: support fail-fast?
		jobScope := taskScope.Fork().WithMatrixValues(labels, values)
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

func (r *Runner) handleMixin(ctx context.Context, j manifest.Job, jobScope *scope.Scope) error {
	name := j.Handler.Name
	_, ok := r.jobFile.Mixins[name]
	if !ok {
		return fmt.Errorf("mixin %q doesn't exist", name)
	}

	// TODO: resolve mixin scope from parent
	loc := j.Handler.Location
	r.shell.Reporter.PrintDiagnostics(
		parsetypes.Diagnostics{
			&parsetypes.Diagnostic{
				FileName: loc.FileName,
				Severity: parsetypes.DiagnosticSeverityError,
				Range:    loc.Range,
				Offset:   loc.Offset,
				Err:      fmt.Errorf("only action jobs are supported currently"),
			},
		},
	)

	return errors.New("only action jobs are supported currently")
}

func (r *Runner) handleJob(ctx context.Context, j manifest.Job, jobScope *scope.Scope) error {
	if j.Kind == manifest.JobKindMixin {
		return r.handleMixin(ctx, j, jobScope)
	}

	hResult := r.actionHandlers.GetActionHandler(ctx, j.Handler, ActionParams{
		Logger: r.logger.Named(j.Handler.String()),
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
		JobName:          j.Handler.String(),
		MatrixParameters: jobScope.MatrixValues,
	})

	runCtx, cancelFn := j.WrapContext(ctx)
	defer cancelFn()

	emitter := r.newSignalEmitter(jobScope, j.Hooks)
	return hResult.Handler.HandleAction(runCtx, emitter)
}

func (r *Runner) newSignalEmitter(parentScope *scope.Scope, hooks manifest.SignalHooks) SignalEmitter {
	if len(hooks) == 0 {
		return noopSignalEmitter{}
	}

	return &signalEmitter{
		runner:      r,
		hooks:       hooks,
		parentScope: parentScope,
	}
}

type signalEmitter struct {
	runner      *Runner
	parentScope *scope.Scope
	hooks       manifest.SignalHooks
}

func (em *signalEmitter) EmitSignal(ctx context.Context, name string, data map[string]any) error {
	if err := validateSignalIsAllowed(name); err != nil {
		return err
	}

	jobs, ok := em.hooks[name]
	if !ok {
		em.runner.logger.Debugf("no hooks for signal %q, skip", name)
		return nil
	}

	s := em.parentScope.Fork()
	s.EventData = data

	em.runner.logger.Debugw("call signal", log.NewField("name", name), log.NewField("data", data))
	err := em.runner.runJobGroup(ctx, jobs, s)
	if err != nil {
		// signal handlers are allowed to fail
		em.runner.logger.Warnf("signal %q returned an error: %s", name, err)
	}

	return nil
}

type noopSignalEmitter struct{}

func (_ noopSignalEmitter) EmitSignal(ctx context.Context, name string, data map[string]any) error {
	return validateSignalIsAllowed(name)
}

func validateSignalIsAllowed(name string) error {
	if name == "error" {
		return fmt.Errorf("sending of internal signal is not allowed: %q", name)
	}

	return nil
}
