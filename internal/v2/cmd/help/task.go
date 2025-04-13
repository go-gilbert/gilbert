package help

import (
	_ "embed"
	"text/template"

	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//go:embed resources/task.usage.gohtml
var taskUsageTpl []byte

type TaskUsageInfo struct {
	Palette       cmdutil.UsageColorPalette
	Cmd           *cobra.Command
	WorkflowFlags *pflag.FlagSet
	TaskFlags     *pflag.FlagSet
	GlobalFlags   *pflag.FlagSet
}

func (i *TaskUsageInfo) Heading(str string) string {
	return i.Palette.RenderHeading(str)
}

func (i *TaskUsageInfo) HasTaskFlags() bool {
	return isFlagSetNotEmpty(i.TaskFlags)
}

func (i *TaskUsageInfo) HasWorkflowFlags() bool {
	return isFlagSetNotEmpty(i.WorkflowFlags)
}

func (i *TaskUsageInfo) HasGlobalFlags() bool {
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
		return tpl.Execute(cmd.OutOrStderr(), &info)
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
