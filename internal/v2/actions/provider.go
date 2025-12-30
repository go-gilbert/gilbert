package actions

import (
	"context"

	"github.com/go-gilbert/gilbert/internal/v2/actions/golang"
	"github.com/go-gilbert/gilbert/internal/v2/engine"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
)

var (
	// Provider is handler provider for builtin actions.
	Provider engine.ActionHandlerProvider = ActionHandlerProvider{}

	namespaces = map[string]map[string]engine.ActionHandlerConstructor{
		"go": golang.Actions,
	}
)

type ActionHandlerProvider struct{}

func (p ActionHandlerProvider) GetActionHandler(ctx context.Context, ref manifest.JobHandlerRef, ap engine.ActionParams) engine.HandlerResult {
	ns, ok := namespaces[ref.Namespace]
	if !ok {
		return engine.NewHandlerNotFoundResult(ref)
	}

	ctor, ok := ns[ref.Name]
	if !ok {
		return engine.NewHandlerNotFoundResult(ref)
	}

	return ctor(ctx, ref, ap)
}
