package debug

import (
	"context"
	"errors"

	"github.com/go-gilbert/gilbert/internal/engine"
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"

	. "github.com/go-gilbert/gilbert/internal/manifest/argschema"
)

var signalSchema = Struct(
	StringField("signal", func(dst *signalActionArgs, val string) error {
		if val == "" {
			return errors.New("signal name is required")
		}

		dst.signalName = val
		return nil
	}).Required(),
	BoolField("fail", func(dst *signalActionArgs, val bool) error {
		dst.fail = val
		return nil
	}),
)

type signalActionArgs struct {
	signalName string
	fail       bool
}

type signalActionHandler struct {
	logger log.Logger
	args   signalActionArgs
}

func newSignalActionHandler(ctx context.Context, ref manifest.JobHandlerRef, params engine.ActionParams) engine.HandlerResult {
	args, diags := MapArgsToStruct(ctx, params, signalSchema)
	if diags.HasError() {
		return engine.NewBadActionParamsResult(ref, diags)
	}

	h := &signalActionHandler{
		logger: params.Logger,
		args:   args,
	}

	return engine.NewHandlerResult(h, diags)
}

func (h *signalActionHandler) HandleAction(ctx context.Context, emitter engine.SignalEmitter) error {
	if h.args.fail {
		return errors.New("test error from action handler")
	}

	err := emitter.EmitSignal(ctx, h.args.signalName, map[string]any{
		"message": "test signal",
	})

	return err
}
