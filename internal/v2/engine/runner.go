// Package engine implements task run engine and its core functionality.
package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

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
}

type Runner struct {
	logger         log.Logger
	jobFile        *manifest.JobFile
	rootScope      *scope.Scope
	shell          *Shell
	cmdProcBuilder CommandProcessorFactory
	actionHandlers ActionHandlerProvider
}

func NewRunner(cfg Config) *Runner {
	return &Runner{
		logger:         cfg.Logger.Named("runner"),
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
	for _, j := range t.Jobs {
		if err := r.runJob(ctx, j, s); err != nil {
			return err
		}
	}

	return nil
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

	r.handleJob(ctx, j, taskScope)
	return nil
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

		jobScope := taskScope.Fork().WithMatrixValues(labels, values)
		err := r.handleJob(ctx, j, jobScope)
		if err != nil {
			return fmt.Errorf("job %q (%s) returned an error: %w", j.Handler, formatMatParams(jobScope.MatrixValues), err)
		}
	}

	return nil
}

func (r *Runner) handleJob(ctx context.Context, j manifest.Job, jobScope *scope.Scope) error {
	if j.Kind != manifest.JobKindAction {
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

		time.Sleep(j.Delay)
	}

	r.shell.Reporter.OnJobStart(JobStartEvent{
		JobName:          j.Handler.String(),
		MatrixParameters: jobScope.MatrixValues,
	})

	// r.logger.Infow(
	// 	"handleJob",
	// 	log.NewField("args", parsetypes.Spew(j.Args).String()),
	// )

	return hResult.Handler.HandleAction(ctx)
}
