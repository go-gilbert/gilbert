package cmd

import (
	"context"

	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/runner"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/expr"
)

type taskRunConfig struct {
	taskName     string
	logger       *log.Logger
	jobFile      *manifest.JobFile
	scope        *scope.Scope
	bootstapOpts cmdutil.BootstrapOpts
}

func startTaskRunner(ctx context.Context, cfg taskRunConfig) error {
	r := runner.NewRunner(runner.Config{
		Logger:    cfg.logger,
		JobFile:   cfg.jobFile,
		RootScope: cfg.scope.Root,
		CmdProcessorFactory: func(s *scope.Scope) expr.CommandProcessor {
			return scope.NewCommandRunner(s)
		},
	})

	err := r.RunTaskWithScope(ctx, cfg.taskName, cfg.scope)
	return err
}
