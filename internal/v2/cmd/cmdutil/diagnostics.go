package cmdutil

import (
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/ui"
	"github.com/go-gilbert/gilbert/internal/v2/ui/theme"
)

func NewDiagnosticsRenderer(logger *log.Logger, opts BootstrapOpts) ui.DiagnosticRenderer {
	if opts.JSON {
		return ui.NewJSONDiagnosticRenderer(logger)
	}

	return ui.NewTerminalDiagnosticsRenderer(log.DefaultIOStreams.Stderr, theme.NewPalette(opts.NoColor))
}
