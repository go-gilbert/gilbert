package log

import "github.com/hashicorp/hcl/v2"

var defaultPrinter = NewOutputPrinter(ErrorLevel, NewStdoutWriter(), TextEncoder{})

type Logger interface {
	Debug(args ...any)
	Debugf(msg string, args ...any)
	Info(args ...any)
	Infof(msg string, args ...any)
	Warn(args ...any)
	Warnf(msg string, args ...any)
	Error(args ...any)
	Errorf(msg string, args ...any)
	Fatal(args ...any)
	Fatalf(msg string, args ...any)
	Panic(args ...any)
	Panicf(msg string, args ...any)
}

type Printer interface {
	Logger
	Success(args ...any)
	Successf(msg string, args ...any)
	Trace(args ...any)
	Tracef(msg string, args ...any)
	ReportDiagnostics(src []byte, diags hcl.Diagnostics)
}

// Global returns default failsafe printer
func Global() Printer {
	return defaultPrinter
}
