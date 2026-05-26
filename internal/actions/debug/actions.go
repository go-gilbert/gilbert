// Package debug provides actions for debugging the runner.
package debug

import (
	"github.com/go-gilbert/gilbert/internal/engine"
)

var Actions = map[string]engine.ActionHandlerConstructor{
	"signal": newSignalActionHandler,
	"echo":   newEchoActionHandler,
}
