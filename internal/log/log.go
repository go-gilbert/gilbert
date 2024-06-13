package log

import (
	"io"
)

// Default is default logger instance
var Default Logger

// Logger is logger interface for logging messages
type Logger interface {
	// Format formats a specified message
	Format(format string, args ...any) string

	// Log logs a message
	Log(args ...any)

	// Logf formats and logs a message
	Logf(message string, args ...any)

	// Debug writes a debug message
	Debug(args ...any)

	// Debugf formats and writes a debug message
	Debugf(message string, args ...any)

	// Warn writes a warning message
	Warn(args ...any)

	// Warnf formats and writes a warning message
	Warnf(message string, args ...any)

	// Error writes an error message
	Error(args ...any)

	// Errorf formats and writes an error message
	Errorf(message string, args ...any)

	// Info writes an info level message
	Info(args ...any)

	// Infof formats and writes an info level message
	Infof(message string, args ...any)

	// Success logs an success message
	Success(args ...any)

	// Successf formats and logs an success message
	Successf(message string, args ...any)

	// Write implements io.Writer interface
	Write(data []byte) (int, error)

	// ErrorWriter returns an io.Writer instance.
	//
	// Used for logging errors from StdErr of other processes.
	ErrorWriter() io.Writer
}

// UseConsoleLogger bootstraps console logger as default log instance
func UseConsoleLogger(level int, noColor bool) {
	Default = &logger{
		level:     level,
		formatter: &paddingFormatter{},
		writer:    &consoleWriter{noColor: noColor},
	}
}

func NewConsoleLogger(level int, noColor bool) Logger {
	return &logger{
		level:     level,
		formatter: &paddingFormatter{},
		writer:    &consoleWriter{noColor: noColor},
	}
}
