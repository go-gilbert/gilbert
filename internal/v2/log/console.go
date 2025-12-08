package log

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

var (
	dbgColor     = color.RGB(66, 66, 66).Add(color.ResetBold)
	errColor     = color.New(color.FgHiRed, color.Bold)
	warnColor    = color.New(color.FgHiYellow, color.Bold)
	successColor = color.New(color.FgGreen)
	whiteColor   = color.New(color.FgHiWhite, color.Bold)
	noColor      = color.New(color.Reset)
)

var _ Writer = (*ConsoleWriter)(nil)

type ConsoleWriter struct {
	// NoColor disables ansi colors output.
	NoColor bool
	Stdout  io.Writer
	Stderr  io.Writer
}

// NewConsoleWriter returns a new writer that writes log messages in human-friendly format.
func NewConsoleWriter(streams IOStreams, noColor bool) ConsoleWriter {
	return ConsoleWriter{
		NoColor: noColor,
		Stdout:  streams.Stdout,
		Stderr:  streams.Stderr,
	}
}

func (c ConsoleWriter) writeNoColor(level Level, prefix, message string, fields []Field) {
	dst := writerForLevel(level, c.Stdout, c.Stderr)
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
		prefixColor *color.Color
		textColor   *color.Color
		prefix      string
	)
	switch level {
	case LevelFatal:
		prefix = "fatal error: "
		prefixColor = errColor
		textColor = whiteColor
	case LevelError:
		prefix = "error: "
		prefixColor = errColor
		textColor = whiteColor
	case LevelWarning:
		prefix = "warning: "
		prefixColor = warnColor
		textColor = whiteColor
	case LevelSuccess:
		textColor = successColor
	case LevelDebug:
		prefix = "debug: "
		textColor = dbgColor
	default:
		c.writeNoColor(level, "", message, fields)
		return
	}

	if prefixColor == nil {
		prefixColor = textColor
	}

	dst := writerForLevel(level, c.Stdout, c.Stderr)
	if c.NoColor {
		c.writeNoColor(level, prefix, message, fields)
		return
	}

	if prefix != "" {
		prefixColor.Fprint(dst, prefix)
	}

	textColor.Fprint(dst, message)
	if len(fields) > 0 {
		dbgColor.Fprint(dst, "\t")

		for _, f := range fields {
			dbgColor.Fprintf(dst, " %s=%v", f.Key, f.Value)
		}
	}

	noColor.Fprintln(dst)
}

func writerForLevel(level Level, stdout, stderr io.Writer) io.Writer {
	switch level {
	case LevelFatal, LevelError, LevelWarning, LevelDebug:
		return stderr
	default:
		return stdout
	}
}
