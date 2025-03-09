package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/spf13/cobra"
)

func newCmdRun(opts RunOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <task> [flags]",
		Short: "Run a task defined in " + cmdutil.DefaultWorkflowFilename,

		// TODO: fill docs someday
		Long: heredoc.Docf(`
			Run a task defined in "tasks" section of %s file.
			Each task may accept a number of input parameters defined in "inputs" section.
		`, cmdutil.DefaultWorkflowFilename),

		RunE: func(_ *cobra.Command, args []string) error {
			// This handler will be executed when task doesn't exist or workflow file has errors.
			return handleTaskNotFound(opts, args)
		},
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if opts.Workflow == nil {
				return
			}

			if len(opts.Workflow.Diagnostics) > 0 {
				cmdutil.RenderDiagnostics(opts.Logger, opts.GlobalDefaults, opts.Workflow.Diagnostics)
			}
		},
	}

	if opts.Workflow != nil && !opts.Workflow.HasErrors {
		addTaskCommands(cmd, opts.Workflow.File)
	}

	return cmd
}

func addTaskCommands(dst *cobra.Command, jf manifest.JobFile) {
	for name, task := range jf.Tasks {
		shortDoc, longDoc := getTaskDescription(name, task)
		cmd := &cobra.Command{
			Use:   name + " [flags]",
			Short: shortDoc,
			Long:  longDoc,
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd:   true,
				DisableNoDescFlag:   true,
				DisableDescriptions: true,
				HiddenDefaultCmd:    true,
			},
		}

		// TODO: mount flags
		dst.AddCommand(cmd)
	}
}

func getTaskDescription(name string, t *manifest.JobGroup) (string, string) {
	switch len(t.Doc) {
	case 0:
		str := fmt.Sprintf("run %q task", name)
		return str, str
	case 1:
		str := t.Doc[0]
		return str, str
	default:
		shortStr := t.Doc[0]
		longStr := strings.Join(t.Doc, "\n")
		return shortStr, longStr
	}
}

func handleTaskNotFound(opts RunOpts, args []string) error {
	if len(args) == 0 {
		return errors.New("task name is required")
	}

	if opts.WorkflowLoadError != nil {
		return fmt.Errorf("workflow file cannot be loaded: %w", opts.WorkflowLoadError)
	}

	if opts.Workflow == nil {
		return fmt.Errorf(
			`no %s file was found in a working directory. Run "gilbert init" to create one`,
			cmdutil.DefaultWorkflowFilename,
		)
	}

	if opts.Workflow.HasErrors {
		return fmt.Errorf("workflow file %q contains errors", opts.Workflow.File.Path)
	}

	return fmt.Errorf("task %q doesn't exist", args[0])
}
