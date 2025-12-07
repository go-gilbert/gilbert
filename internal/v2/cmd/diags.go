package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

// newCmdDiagnostics constructs a command to render workflow file diagnostics (warnings and errors).
func newCmdDiagnostics(opts RunOpts) *cobra.Command {
	diagRenderer := cmdutil.NewDiagnosticsRenderer(opts.Logger, opts.GlobalDefaults)

	var diags parsetypes.Diagnostics
	if opts.Workflow != nil {
		diags = opts.Workflow.Diagnostics
	}

	return &cobra.Command{
		Use:     "diagnostics",
		Short:   "Check workflow file syntax errors",
		Long:    "Check workflow file syntax and display warnings or errors (if any).",
		Aliases: []string{"diags"},
		RunE: func(cmd *cobra.Command, args []string) error {
			diagRenderer.RenderDiagnostics(diags)
			diagRenderer.Reset()

			if len(diags) > 0 {
				os.Exit(1)
			}

			return nil
		},
	}
}

// newCmdList constructs a command to print list of tasks declared in a workflow file.
func newCmdList(opts RunOpts) *cobra.Command {
	diagRenderer := cmdutil.NewDiagnosticsRenderer(opts.Logger, opts.GlobalDefaults)

	return &cobra.Command{
		Use:     "list",
		Short:   "Print list of available tasks",
		Long:    "Check workflow file syntax and display warnings or errors (if any).",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Workflow == nil {
				return fmt.Errorf(`no %q file found. Run "gilbert init" to create a new project`, cmdutil.DefaultWorkflowFilename)
			}

			diagRenderer.RenderDiagnostics(opts.Workflow.Diagnostics)
			diagRenderer.Reset()
			if opts.Workflow.HasErrors {
				return errors.New("workflow file contains errors")
			}

			if opts.GlobalDefaults.JSON {
				renderListJSON(opts.Logger, opts.Workflow.File)
			} else {
				renderListText(opts.GlobalDefaults.NoColor, opts.Workflow.File)
			}

			return nil
		},
	}
}

func renderListJSON(l *log.Logger, jf manifest.JobFile) {
	logger := l.Named("cmd.list.task")
	for name, task := range jf.Tasks {
		logger.Infow(name,
			log.NewField("task", task),
		)
	}
}

func renderListText(noColor bool, jf manifest.JobFile) {
	pal := cmdutil.NewTaskListColorPalette(noColor)
	cmdutil.Println(pal.SectionTitle, "AVAILABLE TASKS")
	cmdutil.Print(pal.Reset)

	maxNameLen := 15
	tasks := make([]*manifest.JobGroup, 0, len(jf.Tasks))
	for name, t := range jf.Tasks {
		maxNameLen = max(maxNameLen, len(name))
		tasks = append(tasks, t)
	}
	padding := strings.Repeat(" ", maxNameLen)

	// Sort tasks by name to keep output order stable.
	slices.SortFunc(tasks, func(a, b *manifest.JobGroup) int {
		return strings.Compare(a.Name, b.Name)
	})

	for _, t := range tasks {
		name := t.Name
		if len(t.Doc) == 0 {
			fmt.Printf("  %s\n", name)
			continue
		}

		lineWrote := false
		fmt.Printf("  %s%s  ", name, padding[len(name):])
		for _, line := range t.Doc {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// TODO
			if !lineWrote {
				fmt.Println(line)
				lineWrote = true
				continue
			}

			fmt.Printf("  %s  %s\n", padding, line)
		}
	}

	cmdutil.Println(pal.Reset)
	fmt.Println(`Use "gilbert run [name] --help" to show information about a task and required parameters.`)
}
