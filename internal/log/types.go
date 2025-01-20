package log

import "io"

// Logger is logger interface for logging messages
type Logger interface {
	// SubLogger creates a new sub-logger
	SubLogger() Logger

	// Format formats a specified message
	Format(format string, args ...interface{}) string

	// Log logs a message
	Log(args ...interface{})

	// Logf formats and logs a message
	Logf(message string, args ...interface{})

	// Debug writes a debug message
	Debug(args ...interface{})

	// Debugf formats and writes a debug message
	Debugf(message string, args ...interface{})

	// Warn writes a warning message
	Warn(args ...interface{})

	// Warnf formats and writes a warning message
	Warnf(message string, args ...interface{})

	// Error writes an error message
	Error(args ...interface{})

	// Errorf formats and writes an error message
	Errorf(message string, args ...interface{})

	// Info writes an info level message
	Info(args ...interface{})

	// Infof formats and writes an info level message
	Infof(message string, args ...interface{})

	// Success logs a success message
	Success(args ...interface{})

	// Successf formats and logs an success message
	Successf(message string, args ...interface{})

	// Write implements io.Writer interface
	Write(data []byte) (int, error)

	// ErrorWriter returns an io.Writer instance.
	//
	// Used for logging errors from StdErr of other processes.
	ErrorWriter() io.Writer
}
