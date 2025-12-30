// Package golang contains action handlers for Go actions.
package golang

import (
	"github.com/go-gilbert/gilbert/internal/v2/engine"
)

var Actions = map[string]engine.ActionHandlerConstructor{
	"build": NewBuildActionHandler,
	"run":   NewRunActionHandler,
}
