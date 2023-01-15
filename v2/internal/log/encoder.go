package log

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"go.uber.org/zap/buffer"
)

var (
	errorColor   = color.New(color.FgRed)
	warnColor    = color.New(color.FgYellow)
	debugColor   = color.New(color.FgCyan)
	successColor = color.New(color.FgGreen)
	noColor      = color.New(color.Reset)

	logPrefix = map[Level]string{
		ErrorLevel: "ERROR:",
		FatalLevel: "FATAL:",
		PanicLevel: "PANIC:",
		WarnLevel:  "Warning:",
		DebugLevel: "Debug:",
	}
)

func EncoderFromString(name string) (Encoder, error) {
	switch name {
	case "json":
		return JSONEncoder{}, nil
	case "text":
		return TextEncoder{}, nil
	case "color":
		return ColorEncoder{}, nil
	}

	return nil, fmt.Errorf("unknown log encoder %q", name)
}

type JSONEncoder struct{}

func (_ JSONEncoder) EncodeEvent(event Event, buff *buffer.Buffer) error {
	err := json.NewEncoder(buff).Encode(event)
	if err != nil {
		return err
	}

	_, err = buff.WriteString("\n")
	return err
}

// ColorEncoder encodes log event as colorful message.
type ColorEncoder struct{}

func (enc ColorEncoder) printWithColor(c *color.Color, event Event, buff *buffer.Buffer) (err error) {
	prefix, ok := logPrefix[event.Level]
	if ok {
		_, err = c.Fprint(buff, prefix, event.Message)
	} else {
		_, err = c.Fprint(buff, prefix)
	}

	if err != nil {
		return err
	}

	_, err = noColor.Print()
	return err
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
		_, err := fmt.Fprint(buff, "::", event.Message)
		return err
	case StyleSuccess:
		return enc.printWithColor(successColor, event, buff)
	}

	_, err := buff.WriteString(event.Message)
	return err
}

type TextEncoder struct{}

func (enc TextEncoder) print(event Event, buff *buffer.Buffer) (err error) {
	prefix := logPrefix[event.Level]
	_, err = fmt.Fprint(buff, prefix, event.Message)

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
		_, err := fmt.Fprint(buff, "::", event.Message)
		return err
	}

	_, err := buff.WriteString(event.Message)
	return err
}
