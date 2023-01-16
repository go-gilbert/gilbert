package log

import "github.com/hashicorp/hcl/v2"

var defaultPrinter Printer = NewOutputPrinter(ErrorLevel, NewStdoutWriter(), TextEncoder{})

// Logger is common abstract logger interface
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

// Printer is high-level output interface for command-line application.
//
// Superset of Logger interface.
type Printer interface {
	Logger

	// Success logs operation success message.
	Success(args ...any)

	// Successf formats and logs operation success message.
	Successf(msg string, args ...any)

	// Trace logs process step message.
	Trace(args ...any)

	// Tracef formats string and logs using Trace.
	Tracef(msg string, args ...any)

	// ReportDiagnostics logs file syntax errors using file source and diagnostics list.
	ReportDiagnostics(diags hcl.Diagnostics)
}

// Global returns default failsafe printer
func Global() Printer {
	return defaultPrinter
}

// ReplaceGlobal replaces global printer with a specified one.
func ReplaceGlobal(p Printer) {
	defaultPrinter = p
}
