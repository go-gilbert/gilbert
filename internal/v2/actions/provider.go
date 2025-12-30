package actions

import (
	"context"
	"log"

	"github.com/go-gilbert/gilbert/internal/v2/engine"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
)

type ActionHandlerProvider struct{}

func (p ActionHandlerProvider) GetActionHandler(ctx context.Context, ref manifest.JobHandlerRef, ap engine.ActionParams) (engine.ActionHandler, error) {
	return nil, nil
}

type GoBuildActionParams struct{}

type GoBuildActionHandler struct {
	logger *log.Logger
	shell  *engine.Shell
}
