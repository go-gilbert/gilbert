package cmdutil

import (
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/ui"
	"github.com/go-gilbert/gilbert/internal/ui/theme"
)

func NewDiagnosticsRenderer(logger *log.Logger, opts BootstrapOpts) ui.DiagnosticRenderer {
	if opts.JSON {
		return ui.NewJSONDiagnosticRenderer(logger)
	}

	return ui.NewTerminalDiagnosticsRenderer(log.DefaultIOStreams.Stderr, theme.NewPalette(opts.NoColor))
}
