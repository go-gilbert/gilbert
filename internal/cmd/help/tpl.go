package help

import (
	"strings"
	"unicode"

	"github.com/go-gilbert/gilbert/pkg/text"
	"github.com/spf13/pflag"
)

const helpIndent = "  "

var tplFuncs = map[string]any{
	"rpad": func(str string, pad int) string {
		padCount := pad - len(str)
		if padCount < 1 {
			return str
		}

		return str + strings.Repeat(" ", padCount)
	},
	"indent": text.Indent,
	"trimTrailingWhitespaces": func(str string) string {
		return strings.TrimRightFunc(str, func(r rune) bool {
			return unicode.IsSpace(r)
		})
	},
	"flagUsages": flagUsages,
}

func flagUsages(fset *pflag.FlagSet) string {
	usages := fset.FlagUsages()
	if usages == "" {
		return usages
	}

	return text.Indent(text.Dedent(usages), "  ")
}
