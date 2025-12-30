package engine

import (
	"context"
	"fmt"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type ActionParams struct {
	Logger log.Logger
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

type HandlerResult struct {
	Handler     ActionHandler
	Diagnostics parsetypes.Diagnostics
	Error       error
}

func NewHandlerResult(h ActionHandler, diags parsetypes.Diagnostics) HandlerResult {
	return HandlerResult{
		Handler:     h,
		Diagnostics: diags,
	}
}

func NewBadActionParamsResult(ref manifest.JobHandlerRef, diags parsetypes.Diagnostics) HandlerResult {
	return HandlerResult{
		Diagnostics: diags,
		Error:       fmt.Errorf("invalid arguments for action %q", ref),
	}
}

func NewHandlerNotFoundResult(ref manifest.JobHandlerRef) HandlerResult {
	loc := ref.Location
	return HandlerResult{
		Error: fmt.Errorf("unknown action %q", ref),
		Diagnostics: parsetypes.Diagnostics{
			&parsetypes.Diagnostic{
				FileName: loc.FileName,
				Severity: parsetypes.DiagnosticSeverityError,
				Range:    loc.Range,
				Offset:   loc.Offset,
				Note:     "action doesn't exists or import is missing",
				Err:      fmt.Errorf("cannot find handler for action"),
			},
		},
	}
}

// ActionHandlerConstructor is action handler constructor function type.
//
// Not used in the engine itself. Provided for convenience of [ActionHandlerProvider] implementors.
type ActionHandlerConstructor = func(ctx context.Context, ref manifest.JobHandlerRef, params ActionParams) HandlerResult

type ActionHandlerProvider interface {
	// GetActionHandler returns an action handler to handle a job.
	//
	// May return [parsetypes.Diagnostics] as error value to display syntax errors.
	GetActionHandler(ctx context.Context, ref manifest.JobHandlerRef, p ActionParams) HandlerResult
}

type ActionHandler interface {
	HandleAction(ctx context.Context) error
}
