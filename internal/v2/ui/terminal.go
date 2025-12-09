package ui

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/runner"
	"github.com/go-gilbert/gilbert/internal/v2/ui/theme"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

var (
	_ runner.Reporter = (*TerminalReporter)(nil)
	_ runner.Input    = (*TerminalInput)(nil)
)

// NewTerminalShell constructs a new terminal shell.
func NewTerminalShell(streams log.IOStreams, noColor bool) *runner.Shell {
	palette := theme.NewPalette(noColor)

	return &runner.Shell{
		Reporter: NewTerminalReporter(streams, palette),
		Input: &TerminalInput{
			Stdin:   streams.Stdin,
			Stdout:  streams.Stdout,
			Palette: palette,
		},
	}
}

type TerminalReporter struct {
	stdout       io.Writer
	stderr       io.Writer
	palette      theme.Palette
	diagRenderer *TerminalDiagnosticRenderer
}

func NewTerminalReporter(stdio log.IOStreams, palette theme.Palette) *TerminalReporter {
	return &TerminalReporter{
		diagRenderer: NewTerminalDiagnosticsRenderer(stdio.Stderr, palette),
		stdout:       stdio.Stdout,
		stderr:       stdio.Stderr,
		palette:      palette,
	}
}

func (r *TerminalReporter) OnTaskStart(taskName string) {
	r.Printf(r.palette.TextHeading, ":: Running task %q\n", taskName)
}

func (r *TerminalReporter) OnJobStart(e runner.JobStartEvent) {
	b := &bytes.Buffer{}
	r.palette.NoteMarker.Fprint(b, "-> ")
	r.palette.Reset.Fprint(b, e.JobName)
	if len(e.MatrixKeys) > 0 {
		b.WriteString(" (")
		for i, k := range e.MatrixKeys {
			if i != 0 {
				b.WriteByte(' ')
			}

			b.WriteString(k)
			fmt.Fprintf(b, "=%v", e.MatrixValues[i])
		}
		b.WriteString(")")
	}

	b.WriteRune('\n')
	r.stdout.Write(b.Bytes())
}

func (r *TerminalReporter) PrintDiagnostics(diags parsetypes.Diagnostics) {
	r.diagRenderer.RenderDiagnostics(diags)
	r.diagRenderer.Reset()
}

func (r *TerminalReporter) Print(p theme.Color, args ...any) {
	_, _ = p.Fprint(r.stdout, args...)
}

func (r *TerminalReporter) Printf(p theme.Color, format string, args ...any) {
	_, _ = p.Fprintf(r.stdout, format, args...)
	_, _ = r.palette.Reset.Fprint(r.stdout)
}

func (r *TerminalReporter) Println(p theme.Color, args ...any) {
	_, _ = p.Fprint(r.stdout, args...)
	_, _ = r.palette.Reset.Fprintln(r.stdout)
}

func (r *TerminalReporter) Eprintln(p theme.Color, args ...any) {
	_, _ = p.Fprintln(r.stderr, args...)
}

func (r *TerminalReporter) Eprintf(p theme.Color, format string, args ...any) {
	_, _ = p.Fprintf(r.stderr, format, args...)
}

type TerminalInput struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Palette theme.Palette
}

func (d *TerminalInput) PromptString(msg string, isRequired bool) (string, error) {
	return "", errors.New("PromptString: not implemented")
}
