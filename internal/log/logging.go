package log

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
