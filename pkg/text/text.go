package text

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var indentRE = regexp.MustCompile(`(?m)^`)

func Dedent(str string) string {
	// see: https://github.com/cli/cli/blob/408e21ebdddf9cd14289e49135389a6e5125eff4/pkg/cmd/root/help.go#L37
	lines := strings.Split(str, "\n")
	minIndent := -1

	for _, l := range lines {
		if len(l) == 0 {
			continue
		}

		indent := len(l) - len(strings.TrimLeft(l, " "))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return str
	}

	var buf bytes.Buffer
	for _, l := range lines {
		_, _ = fmt.Fprintln(&buf, strings.TrimPrefix(l, strings.Repeat(" ", minIndent)))
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

func Indent(str, indent string) string {
	if len(strings.TrimSpace(str)) == 0 {
		return str
	}

	return indentRE.ReplaceAllLiteralString(str, indent)
}
