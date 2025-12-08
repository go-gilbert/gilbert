package log

import (
	"fmt"
	"io"
	"os"
)

var DefaultIOStreams = IOStreams{
	Stdin:  os.Stdin,
	Stdout: os.Stdout,
	Stderr: os.Stderr,
}

type IOStreams struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Field struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// NewField constructs a new log context field.
func NewField(key string, value any) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

// Writer defines an interface for writing log messages.
type Writer interface {
	// Write logs a message with the given level and tag.
	Write(level Level, tag, message string, fields []Field)
}

// Logger represents a logging instance with a name, log level, and a writer.
type Logger struct {
	name   string
	level  Level
	writer Writer
}

// NewLogger constructs a new logger instance.
func NewLogger(name string, level Level, writer Writer) *Logger {
	return &Logger{
		name:   name,
		level:  level,
		writer: writer,
	}
}

// SetWriter replaces log writer.
func (l *Logger) SetWriter(w Writer) {
	l.writer = w
}

// Named returns a new logger with a given name (tag).
func (l Logger) Named(name string) Logger {
	l.name = name
	return l
}

func (l Logger) print(level Level, args ...any) {
	if l.level < level {
		return
	}

	msg := fmt.Sprint(args...)
	l.writer.Write(level, l.name, msg, nil)
}

func (l Logger) printf(level Level, format string, args ...any) {
	if l.level < level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	l.writer.Write(level, l.name, msg, nil)
}

func (l Logger) printw(level Level, msg string, fields []Field) {
	if l.level < level {
		return
	}

	l.writer.Write(level, l.name, msg, fields)
}

// Fatal logs a critical error message and may cause the program to terminate.
func (l Logger) Fatal(args ...any) {
	l.print(LevelFatal, args...)
	os.Exit(1)
}

// Fatalf logs a critical error message using a formatted string.
func (l Logger) Fatalf(format string, args ...any) {
	l.printf(LevelFatal, format, args...)
	os.Exit(1)
}

// Fatalw logs a critical error message with additional context.
func (l Logger) Fatalw(msg string, fields ...Field) {
	l.printw(LevelFatal, msg, fields)
	os.Exit(1)
}

// Error logs an error message indicating a failure.
func (l Logger) Error(args ...any) {
	l.print(LevelError, args...)
}

// Errorf logs an error message using a formatted string.
func (l Logger) Errorf(format string, args ...any) {
	l.printf(LevelError, format, args...)
}

// Errorw logs an error message with additional context.
func (l Logger) Errorw(msg string, fields ...Field) {
	l.printw(LevelError, msg, fields)
}

// Warn logs a warning message indicating a potential issue.
func (l Logger) Warn(args ...any) {
	l.print(LevelWarning, args...)
}

// Warnf logs a warning message using a formatted string.
func (l Logger) Warnf(format string, args ...any) {
	l.printf(LevelWarning, format, args...)
}

// Warnw logs a warning message with additional context.
func (l Logger) Warnw(msg string, fields ...Field) {
	l.printw(LevelWarning, msg, fields)
}

// Success logs a success message indicating a positive outcome.
func (l Logger) Success(args ...any) {
	l.print(LevelSuccess, args...)
}

// Successf logs a success message using a formatted string.
func (l Logger) Successf(format string, args ...any) {
	l.printf(LevelSuccess, format, args...)
}

// Successw logs a success message with additional context.
func (l Logger) Successw(msg string, fields ...Field) {
	l.printw(LevelSuccess, msg, fields)
}

// Info logs an informational message about the application's state.
func (l Logger) Info(args ...any) {
	l.print(LevelInfo, args...)
}

// Infof logs an informational message using a formatted string.
func (l Logger) Infof(format string, args ...any) {
	l.printf(LevelInfo, format, args...)
}

// Infow logs an informational message with additional context.
func (l Logger) Infow(msg string, fields ...Field) {
	l.printw(LevelInfo, msg, fields)
}

// Debug logs a message for debugging purposes.
func (l Logger) Debug(args ...any) {
	l.print(LevelDebug, args...)
}

// Debugf logs a debug message using a formatted string.
func (l Logger) Debugf(format string, args ...any) {
	l.printf(LevelDebug, format, args...)
}

// Debugw logs a debug message with additional context.
func (l Logger) Debugw(msg string, fields ...Field) {
	l.printw(LevelDebug, msg, fields)
}
