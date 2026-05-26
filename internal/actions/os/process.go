package os

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/go-gilbert/gilbert/internal/engine"
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/manifest/argschema"
	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/go-gilbert/gilbert/pkg/executil"
)

var _ engine.ActionHandler = (*ProcessActionHandler)(nil)

type ProcessActionHandler struct {
	logger  log.Logger
	globals scope.Globals
	args    processActionArgs
}

func NewProcessActionHandler(ctx context.Context, ref manifest.JobHandlerRef, params engine.ActionParams) engine.HandlerResult {
	args, diags := argschema.MapArgsToStruct(ctx, params, processSchema)
	if diags.HasError() {
		return engine.NewBadActionParamsResult(ref, diags)
	}

	h := &ProcessActionHandler{
		logger:  params.Logger,
		globals: params.Scope.Globals,
		args:    args,
	}

	return engine.NewHandlerResult(h, diags)
}

func (h *ProcessActionHandler) HandleAction(ctx context.Context, emitter engine.SignalEmitter) error {
	cmd := exec.CommandContext(ctx, h.args.Path, h.args.Args...)
	cmd.Env = executil.MergeEnv(h.args.Env, h.globals.Env)
	cmd.Dir = h.globals.Project.WorkDir
	if h.args.WorkDir != "" {
		cmd.Dir = h.args.WorkDir
	}

	// ensure that subprocess dies as soon as shell is killed.
	executil.AddProcessGroup(cmd)

	if !h.args.Silent {
		w := h.logger.IOWriter()
		cmd.Stdout = w.Stdout
		cmd.Stderr = w.Stderr
	}

	h.logger.Debugw(
		"starting process",
		log.NewField("cwd", cmd.Dir),
		log.NewField("cmd", cmd.Args),
		log.NewField("env", cmd.Env),
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process %q: %w", h.args.Path, err)
	}

	if err := cmd.Wait(); err != nil {
		if h.args.Silent {
			h.logger.Warn("process error log is hidden as silence option is enabled")
		}

		return fmt.Errorf("process %q returned an error: %w", h.args.Path, err)
	}

	return nil
}
