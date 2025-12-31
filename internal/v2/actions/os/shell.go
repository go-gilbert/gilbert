package os

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/go-gilbert/gilbert/internal/v2/engine"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/argschema"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/executil"
)

var _ engine.ActionHandler = (*ShellActionHandler)(nil)

type ShellActionHandler struct {
	logger  log.Logger
	globals scope.Globals
	args    shellActionArgs
}

func NewShellActionHandler(ctx context.Context, ref manifest.JobHandlerRef, params engine.ActionParams) engine.HandlerResult {
	args, diags := argschema.MapArgsToStruct(ctx, params, shellSchema)
	if diags.HasError() {
		return engine.NewBadActionParamsResult(ref, diags)
	}

	h := &ShellActionHandler{
		logger:  params.Logger,
		globals: params.Scope.Globals,
		args:    args,
	}

	return engine.NewHandlerResult(h, diags)
}

func (s *ShellActionHandler) HandleAction(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, s.args.Shell, s.args.ShellCommand, s.args.Command)
	cmd.Env = executil.MergeEnv(s.args.Env, s.globals.Env)
	cmd.Dir = s.globals.Project.WorkDir
	if s.args.WorkDir != "" {
		cmd.Dir = s.args.WorkDir
	}

	// ensure that subprocess dies as soon as shell is killed.
	executil.AddProcessGroup(cmd)

	if !s.args.Silent {
		w := s.logger.IOWriter()
		cmd.Stdout = w.Stdout
		cmd.Stderr = w.Stderr
	}

	s.logger.Debugw(
		"starting command",
		log.NewField("cwd", cmd.Dir),
		log.NewField("cmd", cmd.Args),
		log.NewField("env", cmd.Env),
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command %q: %w", s.args.Command, err)
	}

	if err := cmd.Wait(); err != nil {
		if s.args.Silent {
			s.logger.Warn("command error log is hidden as silence option is enabled")
		}

		return fmt.Errorf("command %q returned an error: %w", s.args.Command, err)
	}

	return nil
}
