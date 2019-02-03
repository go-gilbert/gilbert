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

	// Sprintf formats and adds padding to specified message
	Sprintf(message string, args ...interface{}) string

	// Log logs a message
	Log(message string, args ...interface{})

	// Debug writes a debug message
	Debug(message string, args ...interface{})

	// Warn writes a warning message
	Warn(message string, args ...interface{})

	// Error writes an error message
	Error(message string, args ...interface{})

	// Info writes an info level message
	Info(message string, args ...interface{})

	// Success logs an success message
	Success(message string, args ...interface{})

	// Write logs a raw slice of bytes
	Write(data []byte) (int, error)

	// ErrorWrites returns an io.Writer instance.
	//
	// Used for logging errors from StdErr of other processes.
	ErrorWriter() io.Writer
}
