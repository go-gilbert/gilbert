package log

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

var (
	dbgColor     = color.RGB(66, 66, 66)
	errColor     = color.New(color.FgRed)
	warnColor    = color.New(color.FgYellow)
	successColor = color.New(color.FgGreen)
)

var _ Writer = (*ConsoleWriter)(nil)

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

func (c ConsoleWriter) writeNoColor(level Level, prefix, message string, fields []Field) {
	dst := writerForLevel(level)
	if prefix != "" {
		_, _ = dst.Write([]byte(prefix))
	}

	_, _ = dst.Write([]byte(message))
	if len(fields) > 0 {
		_, _ = dst.Write([]byte("\t"))
		for _, f := range fields {
			_, _ = fmt.Fprintf(dst, " %s=%v", f.Key, f.Value)
		}
	}
	_, _ = dst.Write([]byte("\n"))
}

func (c ConsoleWriter) Write(level Level, _, message string, fields []Field) {
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
	case LevelDebug:
		prefix = "Debug: "
		textColor = dbgColor
	default:
		c.writeNoColor(level, "", message, fields)
		return
	}

	dst := writerForLevel(level)
	if c.NoColor {
		c.writeNoColor(level, prefix, message, fields)
		return
	}

	if prefix != "" {
		_, _ = textColor.Fprint(dst, prefix)
	}

	if len(fields) > 0 {
		_, _ = textColor.Fprint(dst, message, "\t")

		for _, f := range fields {
			_, _ = textColor.Fprintf(dst, " %s=%v", f.Key, f.Value)
		}

		_, _ = textColor.Fprintln(dst)
		return
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
