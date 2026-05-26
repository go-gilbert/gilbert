package help

import (
	_ "embed"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/go-gilbert/gilbert/internal/ui/theme"
)

//go:embed resources/task.usage.gohtml
var taskUsageTpl []byte

type TaskUsageInfo struct {
	NoColor       bool
	Cmd           *cobra.Command
	WorkflowFlags *pflag.FlagSet
	TaskFlags     *pflag.FlagSet
	GlobalFlags   *pflag.FlagSet
}

type taskUsageData struct {
	TaskUsageInfo
	palette theme.Palette
}

func (i *taskUsageData) Heading(str string) string {
	if i.palette.NoColor {
		return str
	}

	return renderHeading(str, &i.palette)
}

func (i *taskUsageData) HasTaskFlags() bool {
	return isFlagSetNotEmpty(i.TaskFlags)
}

func (i *taskUsageData) HasWorkflowFlags() bool {
	return isFlagSetNotEmpty(i.WorkflowFlags)
}

func (i *taskUsageData) HasGlobalFlags() bool {
	return isFlagSetNotEmpty(i.GlobalFlags)
}

func NewTaskUsageFunc(info TaskUsageInfo) UsageFunc {
	return func(cmd *cobra.Command) error {
		// see github.com/spf13/cobra@v1.9.1/command.go:1937 for default template
		tpl, err := template.New("").Funcs(tplFuncs).Parse(string(taskUsageTpl))
		if err != nil {
			return err
		}

		info.Cmd = cmd
		d := &taskUsageData{
			TaskUsageInfo: info,
			palette:       theme.NewPalette(info.NoColor),
		}

		return tpl.Execute(cmd.OutOrStderr(), d)
	}
}

func isFlagSetNotEmpty(fset *pflag.FlagSet) bool {
	if fset == nil {
		return false
	}

	n := 0
	fset.VisitAll(func(_ *pflag.Flag) {
		n++
	})

	return n > 0
}
