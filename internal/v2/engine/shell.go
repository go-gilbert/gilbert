package engine

import "github.com/go-gilbert/gilbert/pkg/parsetypes"

type Shell struct {
	Reporter Reporter
	Input    Input
}

type JobStartEvent struct {
	JobName          string
	MatrixParameters map[string]any
}

type SignalEvent struct {
	SignalName string
	JobName    string
	Args       map[string]any
}

// Reporter reports application updates to user interface.
type Reporter interface {
	// OnTaskStart reports task execution start event.
	OnTaskStart(taskName string)

	// OnJobStart reports task step execution start event.
	OnJobStart(event JobStartEvent)

	// OnSignal reports when action handler emits a signal.
	OnSignal(event SignalEvent)

	// PrintDiagnostics renders diagnostics.
	PrintDiagnostics(diags parsetypes.Diagnostics)
}

// Input interface provides functionality for prompting user input.
type Input interface {
	// PromptString prompt user to input a text.
	//
	// When isRequired is set to true - returns an error if no input was given.
	PromptString(msg string, isRequired bool) (string, error)
}
