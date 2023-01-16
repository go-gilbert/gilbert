package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"strings"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap/buffer"
)

const linesBufferSize = 5

var (
	errorColor   = color.New(color.FgRed)
	warnColor    = color.New(color.FgYellow)
	debugColor   = color.New(color.FgCyan)
	successColor = color.New(color.FgGreen)
	noColor      = color.New(color.Reset)

	diagSummaryColor    = color.New(color.FgHiWhite)
	diagLineNumberColor = color.New(color.FgCyan)

	logPrefix = map[Level]string{
		ErrorLevel: "ERROR: ",
		FatalLevel: "FATAL: ",
		PanicLevel: "PANIC: ",
		WarnLevel:  "Warning: ",
		DebugLevel: "Debug: ",
	}
)

func EncoderFromString(name string) (Encoder, error) {
	switch name {
	case "json":
		return NewJSONEncoder(), nil
	case "text":
		return NewTextEncoder(), nil
	case "color":
		return NewColorEncoder(), nil
	}

	return nil, fmt.Errorf("unknown log encoder %q", name)
}

type diagnosticRecord struct {
	Event
	Diagnostics hcl.Diagnostics `json:"diagnostics"`
}

type JSONEncoder struct{}

func NewJSONEncoder() JSONEncoder {
	return JSONEncoder{}
}

func (_ JSONEncoder) EncodeEvent(event Event, buff *buffer.Buffer) error {
	err := json.NewEncoder(buff).Encode(event)
	if err != nil {
		return err
	}

	_, err = buff.WriteString("\n")
	return err
}

func (enc JSONEncoder) EncodeDiagnostics(diags hcl.Diagnostics, buff *buffer.Buffer) error {
	err := json.NewEncoder(buff).Encode(diagnosticRecord{
		Event: Event{
			Level: ErrorLevel,
			Time:  time.Now(),
		},
		Diagnostics: diags,
	})

	if err != nil {
		return err
	}

	_, err = buff.WriteString("\n")
	return err
}

// ColorEncoder encodes log event as colorful message.
type ColorEncoder struct {
	buff sourceFilesPool
}

func NewColorEncoder() ColorEncoder {
	return ColorEncoder{
		buff: newSourceFilesPool(),
	}
}

func (enc ColorEncoder) printWithColor(c *color.Color, event Event, buff *buffer.Buffer) (err error) {
	prefix, ok := logPrefix[event.Level]
	if ok {
		_, err = c.Fprint(buff, prefix, event.Message)
	} else {
		_, err = c.Fprint(buff, event.Message)
	}

	if err != nil {
		return err
	}

	_, err = noColor.Fprintln(buff)
	return err
}

func (enc ColorEncoder) EncodeDiagnostics(diags hcl.Diagnostics, buff *buffer.Buffer) error {
	// Temporary lines cache for each file
	fileLines := make(map[string][][]byte, linesBufferSize)

	for _, diag := range diags {
		diagColor := errorColor
		if diag.Severity == hcl.DiagWarning {
			diagColor = warnColor
			warnColor.Fprint(buff, "warning")
		} else {
			errorColor.Fprint(buff, "error")
		}

		subj := diag.Subject
		rng := diag.Context
		if rng == nil {
			rng = diag.Subject
		}

		isSingleLine := subj.Start.Line == subj.End.Line
		msg := diag.Summary
		if !isSingleLine {
			msg = diag.Detail
		}

		diagSummaryColor.Fprintln(buff, ":", msg)
		diagLineNumberColor.Fprint(buff, " --> ")

		noColor.Fprintf(buff, "%s:%d:%d\n",
			diag.Subject.Filename, diag.Subject.Start.Line, diag.Subject.Start.Column)

		// Attempt to read file and split by lines
		lines, ok := fileLines[diag.Subject.Filename]
		if !ok {
			data, err := enc.buff.get(diag.Subject.Filename)
			if err != nil {
				errorColor.Fprintln(buff, " *** ERROR: failed to read source file:", err)
				continue
			}

			lines = bytes.Split(data, []byte{'\n'})
			fileLines[diag.Subject.Filename] = lines
		}

		for line := rng.Start.Line; line <= rng.End.Line; line++ {
			text := lines[line-1]
			diagLineNumberColor.Fprintf(buff, "%-3d| ", line)
			noColor.Fprintln(buff, string(text))
			if isSingleLine && line == subj.Start.Line {
				diagLineNumberColor.Fprint(buff, "   | ",
					strings.Repeat(" ", subj.Start.Column-1))
				diagColor.Fprint(buff,
					strings.Repeat("^", subj.End.Column-subj.Start.Column), " ",
					diag.Detail)
				noColor.Fprintln(buff)
			}

		}
		noColor.Fprintln(buff)
	}

	return nil
}

func (enc ColorEncoder) EncodeEvent(event Event, buff *buffer.Buffer) error {
	switch event.Level {
	case PanicLevel, FatalLevel, ErrorLevel:
		return enc.printWithColor(errorColor, event, buff)
	case WarnLevel:
		return enc.printWithColor(warnColor, event, buff)
	case DebugLevel:
		return enc.printWithColor(debugColor, event, buff)
	}

	switch event.Style {
	case StyleStep:
		_, err := fmt.Fprintln(buff, ":: ", event.Message)
		return err
	case StyleSuccess:
		return enc.printWithColor(successColor, event, buff)
	}

	_, err := fmt.Fprintln(buff, event.Message)
	return err
}

type TextEncoder struct {
	buff sourceFilesPool
}

func NewTextEncoder() TextEncoder {
	return TextEncoder{
		buff: newSourceFilesPool(),
	}
}

func (enc TextEncoder) print(event Event, buff *buffer.Buffer) (err error) {
	prefix := logPrefix[event.Level]
	_, err = fmt.Fprintln(buff, prefix, event.Message)

	return err
}

func (enc TextEncoder) EncodeEvent(event Event, buff *buffer.Buffer) error {
	switch event.Level {
	case PanicLevel, FatalLevel, ErrorLevel:
		return enc.print(event, buff)
	case WarnLevel:
		return enc.print(event, buff)
	case DebugLevel:
		return enc.print(event, buff)
	}

	switch event.Style {
	case StyleStep:
		_, err := fmt.Fprintln(buff, ":: ", event.Message)
		return err
	}

	_, err := fmt.Fprintln(buff, event.Message)
	return err
}

func (enc TextEncoder) EncodeDiagnostics(diags hcl.Diagnostics, buff *buffer.Buffer) error {
	// Temporary lines cache for each file
	fileLines := make(map[string][][]byte, linesBufferSize)

	for _, diag := range diags {
		if diag.Severity == hcl.DiagWarning {
			buff.WriteString("warning")
		} else {
			buff.WriteString("error")
		}

		subj := diag.Subject
		rng := diag.Context
		if rng == nil {
			rng = diag.Subject
		}

		isSingleLine := subj.Start.Line == subj.End.Line
		msg := diag.Summary
		if !isSingleLine {
			msg = diag.Detail
		}

		fmt.Fprintln(buff, ":", msg)
		fmt.Fprintf(buff, " --> %s:%d:%d\n",
			diag.Subject.Filename, diag.Subject.Start.Line, diag.Subject.Start.Column)

		// Attempt to read file and split by lines
		lines, ok := fileLines[diag.Subject.Filename]
		if !ok {
			data, err := enc.buff.get(diag.Subject.Filename)
			if err != nil {
				fmt.Fprintln(buff, " *** ERROR: failed to read source file:", err)
				continue
			}

			lines = bytes.Split(data, []byte{'\n'})
			fileLines[diag.Subject.Filename] = lines
		}

		for line := rng.Start.Line; line <= rng.End.Line; line++ {
			text := lines[line-1]
			fmt.Fprintf(buff, "%-3d| ", line)
			fmt.Fprintln(buff, string(text))
			if isSingleLine && line == subj.Start.Line {
				fmt.Fprint(buff, "   | ",
					strings.Repeat(" ", subj.Start.Column-1),
					strings.Repeat("^", subj.End.Column-subj.Start.Column), " ",
					diag.Detail, "\n")
			}

		}
		fmt.Fprintln(buff)
	}

	return nil
}
