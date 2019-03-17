package log

import (
	"fmt"
	"strings"
)

type paddingFormatter struct {
	padding uint
}

func (f *paddingFormatter) Next() Formatter {
	return &paddingFormatter{padding: f.padding + 1}
}

// Format formats log message
func (f *paddingFormatter) Format(format string, args ...interface{}) string {
	return f.WrapString(fmt.Sprintf(format, args...))
}

// WrapStrings wraps string according to current padding
func (f *paddingFormatter) WrapString(str string) string {
	if f.padding == 0 {
		return str
	}
	return strings.Repeat(padChar, int(padding*f.padding)) + str
}
