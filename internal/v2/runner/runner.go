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
	CmdProcessorFactory CommandProcessorFactory

	// TODO: pass hud for TUI
}

type Runner struct {
	logger         log.Logger
	jobFile        *manifest.JobFile
	rootScope      *scope.Scope
	cmdProcBuilder CommandProcessorFactory
}

func NewRunner(cfg Config) *Runner {
	return &Runner{
		logger:         cfg.Logger.Named("runner"),
		jobFile:        cfg.JobFile,
		rootScope:      cfg.RootScope,
		cmdProcBuilder: cfg.CmdProcessorFactory,
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

	for _, j := range t.Jobs {
		r.runJob(ctx, j, s)
	}

	dumpJSON("task", t)

	return fmt.Errorf("not implemented")
}

func (r *Runner) runJob(ctx context.Context, j manifest.Job, taskScope *scope.Scope) error {
	// TODO
	return nil
}

func dumpJSON(pfx string, v any) {
	msg, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println("---", pfx, "---", ":\n", string(msg))
}
