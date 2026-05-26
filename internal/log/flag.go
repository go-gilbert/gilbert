package log

import (
	"fmt"

	"github.com/spf13/pflag"
)

const (
	FormatConsole = "console"
	FormatNoColor = "no-color"
	FormatJSON    = "json"
)

var (
	_ pflag.Value = (*FlagLevel)(nil)
	_ pflag.Value = (*FlagFormat)(nil)
)

type FlagLevel struct {
	dst *Level
}

// NewFlagLevel constructs a new flag value for cobra/pflag to parse value info log level.
func NewFlagLevel(dst *Level) FlagLevel {
	return FlagLevel{
		dst: dst,
	}
}

func (l FlagLevel) String() string {
	if l.dst == nil {
		return ""
	}

	return l.dst.String()
}

func (l FlagLevel) Set(s string) error {
	level := ParseLevel(s)
	if level == LevelUnknown {
		return fmt.Errorf(
			"invalid log level. Supported values: %s, %s, %s, %s, %s, %s",
			LevelFatal,
			LevelError,
			LevelWarning,
			LevelInfo,
			LevelSuccess,
			LevelDebug,
		)
	}

	*l.dst = level
	return nil
}

func (l FlagLevel) Type() string {
	return "string"
}

var errInvalidFlagFormatErr = fmt.Errorf(
	"invalid log format. Allowed values: %s, %s, %s",
	FormatConsole, FormatNoColor, FormatJSON,
)

type FlagFormat struct {
	dst          *Logger
	defaultValue string
}

// NewFlagFormat returns a new flag value to set log writer for pflag/cobra.
func NewFlagFormat(dst *Logger, defaultValue string) FlagFormat {
	return FlagFormat{
		dst:          dst,
		defaultValue: defaultValue,
	}
}

func (l FlagFormat) String() string {
	return l.defaultValue
}

func (l FlagFormat) Set(s string) error {
	if l.dst == nil {
		switch s {
		case FormatJSON, FormatConsole, FormatNoColor:
			return nil
		}

		return errInvalidFlagFormatErr
	}

	var writer Writer
	switch s {
	case FormatJSON:
		writer = NewJSONWriter(DefaultIOStreams)
	case FormatConsole:
		writer = NewConsoleWriter(DefaultIOStreams, false)
	case FormatNoColor:
		writer = NewConsoleWriter(DefaultIOStreams, true)
	default:
		return errInvalidFlagFormatErr
	}

	l.dst.SetWriter(writer)
	return nil
}

func (l FlagFormat) Type() string {
	return "string"
}
