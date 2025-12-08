// Package ui implements user interface functionality.
package ui

type Shell struct {
	Reporter Reporter
	Input    Input
}

// Reporter reports application updates to user interface.
type Reporter interface {
	OnTaskStart(taskName string)
}

// Input interface provides functionality for prompting user input.
type Input interface {
	// PromptString prompt user to input a text.
	//
	// When isRequired is set to true - returns an error if no input was given.
	PromptString(msg string, isRequired bool) (string, error)
}
