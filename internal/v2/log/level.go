package log

import (
	"fmt"
	"strings"
)

type Level uint

func (l Level) String() string {
	switch l {
	case LevelFatal:
		return "fatal"
	case LevelError:
		return "error"
	case LevelWarning:
		return "warn"
	case LevelInfo:
		return "info"
	case LevelSuccess:
		return "success"
	case LevelDebug:
		return "debug"
	default:
		return ""
	}

}

func (l Level) GoString() string {
	switch l {
	case LevelFatal:
		return "LevelFatal"
	case LevelError:
		return "LevelError"
	case LevelWarning:
		return "LevelWarning"
	case LevelInfo:
		return "LevelInfo"
	case LevelSuccess:
		return "LevelSuccess"
	case LevelDebug:
		return "LevelDebug"
	default:
		return fmt.Sprintf("Level(%x)", l)
	}
}

func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

const (
	LevelUnknown Level = iota
	LevelFatal
	LevelError
	LevelWarning
	LevelSuccess
	LevelInfo
	LevelDebug
)

// ParseLevel parses log level from string representation.
//
// Returns LevelUnknown on failure.
func ParseLevel(str string) Level {
	str = strings.ToLower(str)
	switch str {
	case "fatal":
		return LevelFatal
	case "error":
		return LevelError
	case "warn":
		return LevelWarning
	case "success":
		return LevelSuccess
	case "info":
		return LevelInfo
	case "debug":
		return LevelDebug
	default:
		return LevelUnknown
	}
}
