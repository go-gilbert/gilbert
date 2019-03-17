package log

import (
	"fmt"
	"github.com/fatih/color"
)

const (
	padChar = " "

	// padding is default padding for each log level in consoleWriter
	padding = uint(2)
)

// consoleLogWriter is console logger
type consoleWriter struct {
	level int
}

func (c *consoleWriter) Write(level int, message string) {
	switch level {
	case LevelInfo:
		color.Blue(message)
	case LevelSuccess:
		color.Green(message)
	case LevelDebug:
		color.Cyan(message)
	case LevelWarn:
		color.Yellow(message)
	case LevelError:
		color.Red(message)
	default:
		fmt.Print(message)
	}
}
