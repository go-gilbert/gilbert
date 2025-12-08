package ui

import (
	"errors"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
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

func (l *LogReporter) OnJobStart(jobName string, matKeys []string, matValues []any) {
	fields := []log.Field{
		log.NewField("job", jobName),
	}

	if len(matKeys) > 0 {
		m := make(map[string]any, len(matKeys))
		for i, k := range matKeys {
			m[k] = matValues[i]
		}

		fields = append(fields, log.NewField("matrix", m))
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
