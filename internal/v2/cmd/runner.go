package cmd

import (
	"github.com/spf13/cobra"

	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/v2/engine"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/internal/v2/ui"
	"github.com/go-gilbert/gilbert/pkg/expr"
)

type taskRunConfig struct {
	taskName     string
	logger       *log.Logger
	jobFile      *manifest.JobFile
	scope        *scope.Scope
	bootstapOpts cmdutil.BootstrapOpts
}

func startTaskRunner(cmd *cobra.Command, cfg taskRunConfig) error {
	ctx := cmd.Context()

	r := engine.NewRunner(engine.Config{
		Logger:    cfg.logger,
		JobFile:   cfg.jobFile,
		RootScope: cfg.scope.Root,
		Shell:     createShell(cmd, cfg.logger, &cfg.bootstapOpts),
		CmdProcessorFactory: func(s *scope.Scope) expr.CommandProcessor {
			return scope.NewCommandRunner(s)
		},
	})

	err := r.RunTaskWithScope(ctx, cfg.taskName, cfg.scope)
	return err
}

func createShell(cmd *cobra.Command, l *log.Logger, opts *cmdutil.BootstrapOpts) *engine.Shell {
	if opts.JSON {
		return ui.NewHeadlessShell(l)
	}

	stdio := log.IOStreams{
		Stdin:  cmd.InOrStdin(),
		Stdout: cmd.OutOrStdout(),
		Stderr: cmd.OutOrStderr(),
	}

	return ui.NewTerminalShell(stdio, opts.NoColor)
}
