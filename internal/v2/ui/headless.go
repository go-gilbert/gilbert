package ui

import (
	"errors"

	"github.com/go-gilbert/gilbert/internal/v2/engine"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

var (
	_ engine.Reporter = (*LogReporter)(nil)
	_ engine.Input    = (*NoOpInput)(nil)
)

// NewHeadlessShell constructs a new headless shell.
func NewHeadlessShell(l *log.Logger) *engine.Shell {
	return &engine.Shell{
		Reporter: NewLogReporter(l),
		Input:    &NoOpInput{},
	}
}

type LogReporter struct {
	log          log.Logger
	diagRenderer *JSONDiagnosticRenderer
}

// NewLogReporter returns a dummy [Reporter] that writes messages to a log.
func NewLogReporter(l *log.Logger) *LogReporter {
	return &LogReporter{
		log:          l.Named("ui"),
		diagRenderer: NewJSONDiagnosticRenderer(l),
	}
}

func (l *LogReporter) OnJobStart(e engine.JobStartEvent) {
	fields := []log.Field{
		log.NewField("job", e.JobName),
	}

	if len(e.MatrixParameters) > 0 {
		fields = append(fields, log.NewField("matrix", e.MatrixParameters))
	}

	l.log.Infow("starting job", fields...)
}

func (l *LogReporter) OnTaskStart(taskName string) {
	l.log.Infof("starting task %q", taskName)
}

func (l *LogReporter) PrintDiagnostics(diags parsetypes.Diagnostics) {
	l.diagRenderer.RenderDiagnostics(diags)
	l.diagRenderer.Reset()
}

var errInputDisabled = errors.New("user input is disabled")

// NoOpInput is a dummy input implementation.
type NoOpInput struct{}

func (d *NoOpInput) PromptString(msg string, isRequired bool) (string, error) {
	return "", errInputDisabled
}
