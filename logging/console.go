package logging

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

const (
	padChar        = " "
	DefaultPadding = 2
)

type ConsoleErrorWriter struct {
	parent *ConsoleLogger
}

func (w *ConsoleErrorWriter) Write(d []byte) (int, error) {
	color.Red(w.parent.pad(string(d)))
	return len(d), nil
}

type ConsoleLogger struct {
	indent int
	level  int
	debug  bool
}

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

func (c *ConsoleLogger) Sprintf(message string, args ...interface{}) string {
	return c.pad(fmt.Sprintf(message, args...) + "\n")
}

func (c *ConsoleLogger) Log(message string, args ...interface{}) {
	fmt.Print(c.Sprintf(message, args...))
}

func (c *ConsoleLogger) Debug(message string, args ...interface{}) {
	if !c.debug {
		return
	}
	color.Cyan(c.Sprintf(message, args...))
}

func (c *ConsoleLogger) Warn(message string, args ...interface{}) {
	color.Yellow(c.Sprintf(message, args...))
}

func (c *ConsoleLogger) Error(message string, args ...interface{}) {
	color.Red(c.Sprintf(message, args...))
}

func (c *ConsoleLogger) Info(message string, args ...interface{}) {
	color.Blue(c.Sprintf(message, args...))
}

func (c *ConsoleLogger) Success(message string, args ...interface{}) {
	color.Green(c.Sprintf(message, args...))
}

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

func (c *ConsoleLogger) ErrorWriter() io.Writer {
	return &ConsoleErrorWriter{parent: c}
}

func NewConsoleLogger(padding int, showDebugMessages bool) *ConsoleLogger {
	return &ConsoleLogger{
		indent: padding,
		debug:  showDebugMessages,
	}
}
