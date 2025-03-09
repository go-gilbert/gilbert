package cmd

import (
	"errors"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/spf13/cobra"
)

func newCmdRun(opts RunOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <task> [flags]",
		Short: "Run a following task defined in " + cmdutil.DefaultWorkflowFilename,

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

	//cmd.AddCommand(&cobra.Command{
	//	Use: "foo",
	//	Run: func(_ *cobra.Command, _ []string) {
	//		opts.Logger.Infof("opts: %#v", opts.GlobalDefaults)
	//	},
	//})

	return cmd
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

	return fmt.Errorf("task %q is not defined", args[0])
}
