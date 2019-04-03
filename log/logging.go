package log

import "io"

const (
	// LevelMsg is generic message log level
	LevelMsg = iota

	// LevelError is errors log level
	LevelError

	// LevelSuccess is success log level
	LevelSuccess

	// LevelWarn is warning log level
	LevelWarn

	// LevelInfo is info message log level
	LevelInfo

	// LevelDebug is debug messages log level
	LevelDebug
)

// Formatter formats log messages
type Formatter interface {
	// Next returns a new instance of formatter for sub-logger
	Next() Formatter

	// Format formats log message
	Format(format string, args ...interface{}) string

	// WrapString wraps log string
	WrapString(str string) string

	// WrapMultiline wraps multiline string
	WrapMultiline(str string) (out string)
}

// Writer is log writer
type Writer interface {
	// Write writes a message to log with specified level
	Write(level int, message string)
}

// Logger is logger interface for logging messages
type Logger interface {
	// SubLogger creates a new sublogger
	SubLogger() Logger

	// Formats formats a specified message
	Format(format string, args ...interface{}) string

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
	Success(args ...interface{})

	// Success formats and logs an success message
	Successf(message string, args ...interface{})

	// Write implements io.Writer interface
	Write(data []byte) (int, error)

	// ErrorWrites returns an io.Writer instance.
	//
	// Used for logging errors from StdErr of other processes.
	ErrorWriter() io.Writer
}
