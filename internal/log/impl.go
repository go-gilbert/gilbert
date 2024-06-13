package log

import (
	"fmt"
	sdk "github.com/go-gilbert/gilbert-sdk"
	"io"
)

// errorWriter implements io.Writer and uses to write errors from stderr
type errorWriter struct {
	formatter Formatter
	writer    Writer
}

// Write writes raw contents
func (w *errorWriter) Write(d []byte) (int, error) {
	// Trim line break from command line output
	s := w.formatter.WrapMultiline(string(d))
	w.writer.Write(LevelError, s)
	return len(d), nil
}

// consoleLogWriter is console logger
type logger struct {
	level     int
	formatter Formatter
	writer    Writer
}

func (c *logger) log(level int, args ...interface{}) {
	if level > c.level {
		return
	}

	c.writer.Write(level, c.formatter.WrapString(fmt.Sprint(args...))+lineBreak)
}

func (c *logger) logf(level int, format string, args ...interface{}) {
	if level > c.level {
		return
	}

	c.writer.Write(level, c.formatter.Format(format, args...)+lineBreak)
}

func (c *logger) SubLogger() sdk.Logger {
	return &logger{
		level:     c.level,
		formatter: c.formatter.Next(),
		writer:    c.writer,
	}
}

func (c *logger) Format(format string, args ...interface{}) string {
	return c.formatter.Format(format, args...)
}

func (c *logger) Log(args ...interface{}) {
	c.log(LevelMsg, args...)
}

func (c *logger) Logf(format string, args ...interface{}) {
	c.logf(LevelMsg, format, args...)
}

func (c *logger) Debug(args ...interface{}) {
	c.log(LevelDebug, args...)
}

func (c *logger) Debugf(format string, args ...interface{}) {
	c.logf(LevelDebug, format, args...)
}

func (c *logger) Warn(args ...interface{}) {
	c.log(LevelWarn, args...)
}

func (c *logger) Warnf(format string, args ...interface{}) {
	c.logf(LevelWarn, format, args...)
}

func (c *logger) Info(args ...interface{}) {
	c.log(LevelInfo, args...)
}

func (c *logger) Infof(format string, args ...interface{}) {
	c.logf(LevelInfo, format, args...)
}

func (c *logger) Success(args ...interface{}) {
	c.log(LevelSuccess, args...)
}

func (c *logger) Successf(format string, args ...interface{}) {
	c.logf(LevelSuccess, format, args...)
}

func (c *logger) Error(args ...interface{}) {
	c.log(LevelError, args...)
}

func (c *logger) Errorf(format string, args ...interface{}) {
	c.logf(LevelError, format, args...)
}

func (c *logger) Write(data []byte) (int, error) {
	lines := c.formatter.WrapMultiline(string(data))
	c.writer.Write(LevelMsg, lines)

	return len(data), nil
}

func (c *logger) ErrorWriter() io.Writer {
	return &errorWriter{writer: c.writer, formatter: c.formatter}
}
