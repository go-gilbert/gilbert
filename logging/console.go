package logging

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

const (
	padChar = " "

	// DefaultPadding is default padding for each log level in ConsoleLogger
	DefaultPadding = 2
)

// ConsoleErrorWriter implements io.Writer and uses to write errors from stderr
type ConsoleErrorWriter struct {
	parent *ConsoleLogger
}

// Write writes raw contents
func (w *ConsoleErrorWriter) Write(d []byte) (int, error) {
	color.Red(w.parent.pad(string(d)))
	return len(d), nil
}

// ConsoleLogger is console logger
type ConsoleLogger struct {
	indent int
	level  int
	debug  bool
}

// SubLogger creates a new sublogger
func (c *ConsoleLogger) SubLogger() Logger {
	return &ConsoleLogger{
		indent: c.indent,
		level:  c.level + 1,
		debug:  c.debug,
	}
}

func (c *ConsoleLogger) pad(str string) (padding string) {
	if c.level == 0 {
		return str
	}
	return strings.Repeat(padChar, c.indent*c.level) + str
}

// Sprintf formats and adds padding to specified message
func (c *ConsoleLogger) Sprintf(message string, args ...interface{}) string {
	return c.pad(fmt.Sprintf(message, args...) + "\n")
}

// Log logs a message
func (c *ConsoleLogger) Log(message string, args ...interface{}) {
	fmt.Print(c.Sprintf(message, args...))
}

// Debug writes a debug message
func (c *ConsoleLogger) Debug(message string, args ...interface{}) {
	if !c.debug {
		return
	}
	color.Cyan(c.Sprintf(message, args...))
}

// Warn writes a warning message
func (c *ConsoleLogger) Warn(message string, args ...interface{}) {
	color.Yellow(c.Sprintf(message, args...))
}

// Error writes an error message
func (c *ConsoleLogger) Error(message string, args ...interface{}) {
	color.Red(c.Sprintf(message, args...))
}

// Info writes an info level message
func (c *ConsoleLogger) Info(message string, args ...interface{}) {
	color.Blue(c.Sprintf(message, args...))
}

// Success logs an success message
func (c *ConsoleLogger) Success(message string, args ...interface{}) {
	color.Green(c.Sprintf(message, args...))
}

// Write logs a raw slice of bytes
func (c *ConsoleLogger) Write(data []byte) (int, error) {
	lines := strings.Split(string(data), lineBreak)
	for _, line := range lines {
		if line == "" {
			continue
		}

		fmt.Println(c.pad(line))
	}

	return len(data), nil
}

// ErrorWriter returns an io.Writer instance.
//
// Used for logging errors from StdErr of other processes.
func (c *ConsoleLogger) ErrorWriter() io.Writer {
	return &ConsoleErrorWriter{parent: c}
}

// NewConsoleLogger creates a new instance of ConsoleLogger
func NewConsoleLogger(padding int, showDebugMessages bool) *ConsoleLogger {
	return &ConsoleLogger{
		indent: padding,
		debug:  showDebugMessages,
	}
}
