package ui

import (
	"errors"
	"io"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/ui/theme"
)

var (
	_ Reporter = (*TerminalReporter)(nil)
	_ Input    = (*TerminalInput)(nil)
)

// NewTerminalShell constructs a new terminal shell.
func NewTerminalShell(streams log.IOStreams, noColor bool) *Shell {
	palette := theme.NewPalette(noColor)

	return &Shell{
		Reporter: &TerminalReporter{
			Stdout:  streams.Stdout,
			Stderr:  streams.Stderr,
			Palette: palette,
		},
		Input: &TerminalInput{
			Stdin:   streams.Stdin,
			Stdout:  streams.Stdout,
			Palette: palette,
		},
	}
}

type TerminalReporter struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Palette theme.Palette
}

func (r *TerminalReporter) OnTaskStart(taskName string) {
	// TODO
}

func (r *TerminalReporter) Print(p theme.Color, args ...any) {
	_, _ = p.Fprint(r.Stdout, args...)
}

func (r *TerminalReporter) Printf(p theme.Color, format string, args ...any) {
	_, _ = p.Fprintf(r.Stdout, format, args...)
}

func (r *TerminalReporter) Println(p theme.Color, args ...any) {
	_, _ = p.Fprintln(r.Stdout, args...)
}

func (r *TerminalReporter) Eprintln(p theme.Color, args ...any) {
	_, _ = p.Fprintln(r.Stderr, args...)
}

func (r *TerminalReporter) Eprintf(p theme.Color, format string, args ...any) {
	_, _ = p.Fprintf(r.Stderr, format, args...)
}

type TerminalInput struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Palette theme.Palette
}

func (d *TerminalInput) PromptString(msg string, isRequired bool) (string, error) {
	return "", errors.New("PromptString: not implemented")
}
