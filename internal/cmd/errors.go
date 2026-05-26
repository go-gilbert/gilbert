package cmd

import (
	"fmt"

	"github.com/go-gilbert/gilbert/internal/cmd/cmdutil"
)

// ErrorWithNote is a wrapper error type that is used to display additional note after error has been printed.
type ErrorWithNote struct {
	Message string
	Note    string
}

func (err *ErrorWithNote) Error() string {
	return err.Message
}

var (
	errTaskNameRequired error = &ErrorWithNote{
		Message: "task name required",
		Note:    `use "gilbert list" to show available tasks`,
	}

	errWorkflowNotFound error = &ErrorWithNote{
		Message: fmt.Sprintf("no %s file was found in a working directory", cmdutil.DefaultWorkflowFilename),
		Note:    `run "gilbert init" to create one`,
	}
)

func newErrWorkflowFileHasErrors(fp string) error {
	return &ErrorWithNote{
		Message: fmt.Sprintf("workflow file %q contains errors", fp),
		Note:    `use "gilbert diagnostics" to print syntax errors`,
	}
}

func newErrTaskNotFound(name string) error {
	return &ErrorWithNote{
		Message: fmt.Sprintf("task %q doesn't exist", name),
		Note:    `use "gilbert list" to show available tasks`,
	}
}
