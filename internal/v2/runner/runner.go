// Package runner implements task run engine.
package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/expr"
)

type CommandProcessorFactory = func(*scope.Scope) expr.CommandProcessor

type Config struct {
	Logger              *log.Logger
	JobFile             *manifest.JobFile
	RootScope           *scope.Scope
	Shell               *Shell
	CmdProcessorFactory CommandProcessorFactory
}

type Runner struct {
	logger         log.Logger
	jobFile        *manifest.JobFile
	rootScope      *scope.Scope
	shell          *Shell
	cmdProcBuilder CommandProcessorFactory
}

func NewRunner(cfg Config) *Runner {
	return &Runner{
		logger:         cfg.Logger.Named("runner"),
		jobFile:        cfg.JobFile,
		rootScope:      cfg.RootScope,
		cmdProcBuilder: cfg.CmdProcessorFactory,
		shell:          cfg.Shell,
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

	// dumpJSON("task", t)

	return fmt.Errorf("not implemented")
}

func (r *Runner) runJob(ctx context.Context, j manifest.Job, taskScope *scope.Scope) error {
	if len(j.Strategy.Matrix) != 0 {
		return r.runJobMatrix(ctx, j, taskScope)
	}

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

	// do catersian product and run each job
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

		r.shell.Reporter.OnJobStart(JobStartEvent{
			JobName:      j.Handler.String(),
			MatrixKeys:   labels,
			MatrixValues: values,
		})
	}

	return nil
}

func dumpJSON(pfx string, v any) {
	msg, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println("---", pfx, "---", ":\n", string(msg))
}
