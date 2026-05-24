package golang

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

var _ engine.ActionHandler = (*RunActionHandler)(nil)

type RunActionHandler struct {
	logger log.Logger
	shell  *engine.Shell

	globals scope.Globals
	args    runActionArgs
}

func NewRunActionHandler(ctx context.Context, ref manifest.JobHandlerRef, params engine.ActionParams) engine.HandlerResult {
	args, diags := argschema.MapArgsToStruct(ctx, params, runSchema)
	if diags.HasError() {
		return engine.NewBadActionParamsResult(ref, diags)
	}

	h := &RunActionHandler{
		logger:  params.Logger,
		shell:   params.Shell,
		globals: params.Scope.Globals,
		args:    args,
	}

	return engine.NewHandlerResult(h, diags)
}

func (b *RunActionHandler) HandleAction(ctx context.Context, emitter engine.SignalEmitter) error {
	argv, err := b.args.commandArgs()
	if err != nil {
		return err
	}

	env := b.args.buildEnv(b.globals.Env)
	cmd := exec.CommandContext(ctx, "go", argv...)
	cmd.Env = env
	cmd.Dir = b.globals.Project.WorkDir

	w := b.logger.IOWriter()
	cmd.Stdout = w.Stdout
	cmd.Stderr = w.Stderr

	b.logger.Debugw(
		"starting command",
		log.NewField("cwd", cmd.Dir),
		log.NewField("cmd", cmd.Args),
		log.NewField("env", cmd.Env),
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start go command: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return executil.FormatExitError(err)
	}

	return nil
}
