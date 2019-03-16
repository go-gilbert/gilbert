package logging

import "io"

var (
	// Log is a global logger instance
	Log Logger
)

// Logger is logger interface for logging messages
type Logger interface {
	// SubLogger creates a new sublogger
	SubLogger() Logger

	// Formats formats a specified message
	Format(message string, args ...interface{}) string

	// Log logs a message
	Log(args ...interface{})

	// Log formats and logs a message
	Logf(message string, args ...interface{})

	// Debug writes a debug message
	Debug(args ...interface{})

	// Debugf formats and writes a debug message
	Debugf(message string, args ...interface{})

	// Warn writes a warning message
	Warn(args ...interface{})

	// Warn formats and writes a warning message
	Warnf(message string, args ...interface{})

	// Error writes an error message
	Error(args ...interface{})

	// Errorf formats and writes an error message
	Errorf(message string, args ...interface{})

	// Info writes an info level message
	Info(args ...interface{})

	// Infof formats and writes an info level message
	Infof(message string, args ...interface{})

	// Success logs an success message
	Success(message string, args ...interface{})

	// Write implements io.Writer interface
	Write(data []byte) (int, error)

	// ErrorWrites returns an io.Writer instance.
	//
	// Used for logging errors from StdErr of other processes.
	ErrorWriter() io.Writer
}
