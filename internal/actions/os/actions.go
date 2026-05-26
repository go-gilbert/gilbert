// Package os contains system-specific action handlers like running external processes, etc.
package os

import (
	"github.com/go-gilbert/gilbert/internal/engine"
)

var Actions = map[string]engine.ActionHandlerConstructor{
	"shell":   NewShellActionHandler,
	"process": NewProcessActionHandler,
}
