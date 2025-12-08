package ui

import (
	"errors"

	"github.com/go-gilbert/gilbert/internal/v2/log"
)

var (
	_ Reporter = (*LogReporter)(nil)
	_ Input    = (*NoOpInput)(nil)
)

// NewHeadlessShell constructs a new headless shell.
func NewHeadlessShell(l *log.Logger) *Shell {
	return &Shell{
		Reporter: NewLogReporter(l),
		Input:    &NoOpInput{},
	}
}

type LogReporter struct {
	log log.Logger
}

// NewLogReporter returns a dummy [Reporter] that writes messages to a log.
func NewLogReporter(l *log.Logger) *LogReporter {
	return &LogReporter{
		log: l.Named("ui"),
	}
}

func (l *LogReporter) OnTaskStart(taskName string) {
	l.log.Infof("starting task %q", taskName)
}

var errInputDisabled = errors.New("user input is disabled")

// NoOpInput is a dummy input implementation.
type NoOpInput struct{}

func (d *NoOpInput) PromptString(msg string, isRequired bool) (string, error) {
	return "", errInputDisabled
}
