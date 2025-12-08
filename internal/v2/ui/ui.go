// Package ui implements user interface functionality.
package ui

import "github.com/go-gilbert/gilbert/pkg/parsetypes"

type Shell struct {
	Reporter Reporter
	Input    Input
}

// Reporter reports application updates to user interface.
type Reporter interface {
	// OnTaskStart reports task execution start event.
	OnTaskStart(taskName string)

	// OnJobStart reports task step execution start event.
	OnJobStart(stageName string, matKeys []string, matValues []any)

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
