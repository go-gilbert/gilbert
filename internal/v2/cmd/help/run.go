package help

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//go:embed resources/run.usage.gohtml
var runUsageTpl []byte

type RunUsageInfo struct {
	Palette       cmdutil.UsageColorPalette
	Cmd           *cobra.Command
	WorkflowFlags *pflag.FlagSet
	CommandFlags  *pflag.FlagSet
	GlobalFlags   *pflag.FlagSet
}

func (i *RunUsageInfo) Heading(str string) string {
	return i.Palette.RenderHeading(str)
}

func (i *RunUsageInfo) GroupHeading(str string) string {
	str = strings.ToUpper(strings.TrimSuffix(str, ":"))
	return i.Palette.RenderHeading(str)
}

func (i *RunUsageInfo) HasCommandFlags() bool {
	return isFlagSetNotEmpty(i.CommandFlags)
}

func (i *RunUsageInfo) HasWorkflowFlags() bool {
	return isFlagSetNotEmpty(i.WorkflowFlags)
}

func (i *RunUsageInfo) HasGlobalFlags() bool {
	return isFlagSetNotEmpty(i.GlobalFlags)
}

func NewRunUsageFunc(info *RunUsageInfo) UsageFunc {
	return func(cmd *cobra.Command) error {
		tpl, err := template.New("").Funcs(tplFuncs).Parse(string(runUsageTpl))
		if err != nil {
			return err
		}

		info.Cmd = cmd
		return tpl.Execute(cmd.OutOrStderr(), &info)
	}
}
