package log

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	InvalidLevel Level = iota
	FatalLevel
	PanicLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

// Level is log level
type Level int

// IsError returns true if level is ErrorLevel or FatalLevel
func (l Level) IsError() bool {
	return l >= ErrorLevel
}

func (l Level) MarshalText() (text []byte, err error) {
	str := l.String()
	return []byte(str), nil
}

func (l *Level) UnmarshalText(text []byte) error {
	lvl, err := LevelFromString(string(text))
	if err != nil {
		return err
	}

	*l = lvl
	return nil
}

// MarshalJSON implements json.Marshaler
func (l Level) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// String returns level string representation
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	default:
		return ""
	}
}

// LevelFromString parses Level from string
func LevelFromString(str string) (Level, error) {
	switch level := strings.ToLower(str); level {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "panic":
		return PanicLevel, nil
	default:
		return InvalidLevel, fmt.Errorf("invalid log level: %q", str)
	}
}
