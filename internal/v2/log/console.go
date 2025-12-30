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

type noColorMsg struct {
	level   Level
	tag     string
	prefix  string
	message string
	fields  []Field
}

func (c ConsoleWriter) writeNoColor(msg noColorMsg) {
	dst := writerForLevel(msg.level, c.Stdout, c.Stderr)
	if msg.prefix != "" {
		_, _ = dst.Write([]byte(msg.prefix))
	}

	if msg.tag != "" {
		_, _ = fmt.Fprintf(dst, "[%s]: ", msg.tag)
	}

	_, _ = dst.Write([]byte(msg.message))
	if len(msg.fields) > 0 {
		_, _ = dst.Write([]byte("\t"))
		for _, f := range msg.fields {
			_, _ = fmt.Fprintf(dst, " %s=%v", f.Key, f.Value)
		}
	}
	_, _ = dst.Write([]byte("\n"))
}

func (c ConsoleWriter) Write(level Level, tag string, message string, fields []Field) {
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
		c.writeNoColor(noColorMsg{
			level:   level,
			tag:     tag,
			message: message,
			fields:  fields,
		})
		return
	}

	if prefixColor == nil {
		prefixColor = textColor
	}

	dst := writerForLevel(level, c.Stdout, c.Stderr)
	if c.NoColor {
		c.writeNoColor(noColorMsg{
			level:   level,
			tag:     tag,
			prefix:  prefix,
			message: message,
			fields:  fields,
		})
		return
	}

	if prefix != "" {
		prefixColor.Fprint(dst, prefix)
	}

	if tag != "" {
		prefixColor.Fprintf(dst, "[%s] ", tag)
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
