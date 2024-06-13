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

// WrapMultiline wraps multiline string
func (f *paddingFormatter) WrapMultiline(str string) (out string) {
	lines := strings.Split(str, lineBreak)
	for _, line := range lines {
		if line == "" {
			continue
		}

		line = strings.TrimSuffix(line, lineBreak)
		out += f.WrapString(line + lineBreak)
	}

	return out
}

// WrapStrings wraps string according to current padding
func (f *paddingFormatter) WrapString(str string) string {
	if f.padding == 0 {
		return str
	}
	return strings.Repeat(padChar, int(padding*f.padding)) + str
}
