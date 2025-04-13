package help

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/spf13/cobra"
)

//go:embed resources/default.usage.gohtml
var defaultUsageTpl []byte

type generalUsageInfo struct {
	Palette cmdutil.UsageColorPalette
	Cmd     *cobra.Command
}

func (i *generalUsageInfo) Heading(str string) string {
	return i.Palette.RenderHeading(str)
}

func (i *generalUsageInfo) GroupHeading(str string) string {
	str = strings.ToUpper(strings.TrimSuffix(str, ":"))
	return i.Palette.RenderHeading(str)
}

type UsageFunc = func(cmd *cobra.Command) error

func NewGeneralUsageFunc(palette cmdutil.UsageColorPalette) UsageFunc {
	return func(cmd *cobra.Command) error {
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
