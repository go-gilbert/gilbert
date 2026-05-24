package debug

import (
	"context"
	"errors"

	"github.com/go-gilbert/gilbert/internal/v2/engine"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"

	. "github.com/go-gilbert/gilbert/internal/v2/manifest/argschema"
)

var echoSchema = Struct(
	StringField("message", func(dst *echoArgs, val string) error {
		if val == "" {
			return errors.New("message name is required")
		}

		dst.message = val
		return nil
	}).Required(),
)

type echoArgs struct {
	message string
}

type echoActionHandler struct {
	logger log.Logger
	args   echoArgs
}

func newEchoActionHandler(ctx context.Context, ref manifest.JobHandlerRef, params engine.ActionParams) engine.HandlerResult {
	args, diags := MapArgsToStruct(ctx, params, echoSchema)
	if diags.HasError() {
		return engine.NewBadActionParamsResult(ref, diags)
	}

	h := &echoActionHandler{
		logger: params.Logger,
		args:   args,
	}

	return engine.NewHandlerResult(h, diags)
}

func (h *echoActionHandler) HandleAction(ctx context.Context, emitter engine.SignalEmitter) error {
	h.logger.Info(h.args.message)
	return nil
}
