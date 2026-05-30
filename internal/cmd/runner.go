package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/go-gilbert/gilbert/internal/actions"
	"github.com/go-gilbert/gilbert/internal/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/engine"
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/go-gilbert/gilbert/internal/ui"
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
	sh := createShell(cmd, cfg.logger, &cfg.bootstapOpts)

	// Expand expressions in `env` block and append to scope environment vars.
	diags := manifest.AppendLazyEnvVarsToScope(ctx, cfg.scope.Root, cfg.jobFile.Env)
	if len(diags) > 0 {
		sh.Reporter.PrintDiagnostics(diags)
		if diags.HasError() {
			return errors.New("failed to expand environment variables declared in workflow file")
		}
	}

	r := engine.NewRunner(engine.Config{
		Logger:                cfg.logger,
		JobFile:               cfg.jobFile,
		RootScope:             cfg.scope.Root,
		ActionHandlerProvider: actions.Provider,
		Shell:                 sh,
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
