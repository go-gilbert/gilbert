package runner

import (
	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/runner/job"
	"github.com/go-gilbert/gilbert/internal/scope"
)

// ActionHandlers is action handlers map.
//
// Key is an action name and value is action constructor
type ActionHandlers = map[string]HandlerFactory

// HandlerFactory is action handler constructor
type HandlerFactory func(*scope.Scope, manifest.ActionParams) (ActionHandler, error)

// ActionHandler represents Gilbert's action handler
type ActionHandler interface {
	// Call calls an action handler
	Call(ctx *job.RunContext, r *TaskRunner) error

	// Cancel aborts action handler execution
	Cancel(ctx *job.RunContext) error
}
