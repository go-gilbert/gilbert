package help

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/go-gilbert/gilbert/internal/ui/theme"
)

//go:embed resources/default.usage.gohtml
var defaultUsageTpl []byte

type generalUsageInfo struct {
	Palette theme.Palette
	Cmd     *cobra.Command
}

func (i *generalUsageInfo) Heading(str string) string {
	if i.Palette.NoColor {
		return str
	}

	return renderHeading(str, &i.Palette)
}

func (i *generalUsageInfo) GroupHeading(str string) string {
	str = strings.ToUpper(strings.TrimSuffix(str, ":"))
	return i.Heading(str)
}

type UsageFunc = func(cmd *cobra.Command) error

func NewGeneralUsageFunc(noColor bool) UsageFunc {
	return func(cmd *cobra.Command) error {
		palette := theme.NewPalette(noColor)
		tpl, err := template.New("").Funcs(tplFuncs).Parse(string(defaultUsageTpl))
		if err != nil {
			return err
		}

		info := &generalUsageInfo{
			Palette: palette,
			Cmd:     cmd,
		}
		return tpl.Execute(cmd.OutOrStderr(), &info)
	}
}

func renderHeading(str string, pal *theme.Palette) string {
	sb := &strings.Builder{}
	sb.Grow(len(str) + 8)
	pal.TextHeading.Fprint(sb, str)
	pal.Reset.Fprint(sb)
	return sb.String()
}
