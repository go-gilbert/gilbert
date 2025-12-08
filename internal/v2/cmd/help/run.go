package help

import (
	_ "embed"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/go-gilbert/gilbert/internal/v2/ui/theme"
)

//go:embed resources/run.usage.gohtml
var runUsageTpl []byte

type RunUsageInfo struct {
	NoColor       bool
	Cmd           *cobra.Command
	WorkflowFlags *pflag.FlagSet
	CommandFlags  *pflag.FlagSet
	GlobalFlags   *pflag.FlagSet
}

type runUsageData struct {
	RunUsageInfo

	palette theme.Palette
}

func (i *runUsageData) Heading(str string) string {
	if i.NoColor {
		return str
	}

	return renderHeading(str, &i.palette)
}

func (i *runUsageData) GroupHeading(str string) string {
	str = strings.ToUpper(strings.TrimSuffix(str, ":"))
	return i.Heading(str)
}

func (i *runUsageData) HasCommandFlags() bool {
	return isFlagSetNotEmpty(i.CommandFlags)
}

func (i *runUsageData) HasWorkflowFlags() bool {
	return isFlagSetNotEmpty(i.WorkflowFlags)
}

func (i *runUsageData) HasGlobalFlags() bool {
	return isFlagSetNotEmpty(i.GlobalFlags)
}

func NewRunUsageFunc(info *RunUsageInfo) UsageFunc {
	return func(cmd *cobra.Command) error {
		tpl, err := template.New("").Funcs(tplFuncs).Parse(string(runUsageTpl))
		if err != nil {
			return err
		}

		info.Cmd = cmd
		d := &runUsageData{
			RunUsageInfo: *info,
			palette:      theme.NewPalette(info.NoColor),
		}
		return tpl.Execute(cmd.OutOrStderr(), d)
	}
}
