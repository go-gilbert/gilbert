package log

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

var (
	errColor     = color.New(color.FgRed)
	warnColor    = color.New(color.FgYellow)
	successColor = color.New(color.FgGreen)
)

type ConsoleWriter struct {
	// NoColor disables ansi colors output.
	NoColor bool
}

// NewConsoleWriter returns a new writer that writes log messages in human-friendly format.
func NewConsoleWriter(noColor bool) ConsoleWriter {
	return ConsoleWriter{
		NoColor: noColor,
	}
}

func (c ConsoleWriter) writeNoColor(level Level, message string) {
	dst := writerForLevel(level)
	_, _ = fmt.Fprintln(dst, message)
}

func (c ConsoleWriter) Write(level Level, _, message string) {
	if c.NoColor {
		c.writeNoColor(level, message)
		return
	}

	var (
		textColor *color.Color
		prefix    string
	)
	switch level {
	case LevelFatal:
		prefix = "Fatal error: "
		textColor = errColor
	case LevelError:
		prefix = "Error: "
		textColor = errColor
	case LevelWarning:
		prefix = "Warning: "
		textColor = warnColor
	case LevelSuccess:
		textColor = successColor
	default:
		c.writeNoColor(level, message)
		return
	}

	dst := writerForLevel(level)
	if prefix != "" {
		_, _ = textColor.Fprint(dst, prefix)
	}

	_, _ = textColor.Fprintln(dst, message)
}

func writerForLevel(level Level) io.Writer {
	switch level {
	case LevelFatal, LevelError, LevelWarning, LevelDebug:
		return os.Stderr
	default:
		return os.Stdout
	}
}
