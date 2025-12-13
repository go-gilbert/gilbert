package engine

import (
	"context"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/expr"
)

type ActionParams struct {
	Logger *log.Logger
	Shell  *Shell

	// Scope is job execution scope.
	Scope *scope.Scope

	// Args are action call parameters.
	Args manifest.JobArgs

	// EvalParams is expression execution context to be used to resolve lazy-evaluated values in [Args].
	EvalParams expr.EvalParams

	// Outputs is a set of output values that are expected to be returned to a job.
	Outputs []string
}

type ActionHandlerProvider interface {
	// GetActionHandler returns an action handler to handle a job.
	//
	// May return [parsetypes.Diagnostics] as error value to display syntax errors.
	GetActionHandler(ctx context.Context, ref manifest.JobHandlerRef, p ActionParams) (ActionHandler, error)
}

type ActionHandler interface {
	HandleAction(ctx context.Context) error
}
