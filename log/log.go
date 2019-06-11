package log

import "github.com/go-gilbert/gilbert-sdk"

// Default is default logger instance
var Default sdk.Logger

// UseConsoleLogger bootstraps console logger as default log instance
func UseConsoleLogger(level int, noColor bool) {
	Default = &logger{
		level:     level,
		formatter: &paddingFormatter{},
		writer:    &consoleWriter{noColor: noColor},
	}
}
