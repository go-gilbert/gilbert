package cmdutil

import (
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

func RenderDiagnostics(logger *log.Logger, opts BootstrapOpts, diags parsetypes.Diagnostics) {
	if opts.JSON {
		renderDiagnosticsJSON(*logger, diags)
		return
	}

	renderDiagnosticsJSON(*logger, diags)
}

func renderDiagnosticsJSON(logger log.Logger, diags parsetypes.Diagnostics) {
	logger = logger.Named("diagnostics")
	for _, diag := range diags {
		fields := []log.Field{
			log.NewField("file", diag.FileName),
			log.NewField("range", diag.Range),
			log.NewField("offset", diag.Offset),
		}
		if diag.Severity == parsetypes.DiagnosticSeverityWarning {
			logger.Warnw(diag.Err.Error(), fields...)
			continue
		}
		logger.Errorw(diag.Err.Error(), fields...)
	}
}
