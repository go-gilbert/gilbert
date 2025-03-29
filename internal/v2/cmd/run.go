package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil/inputflag"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/spf13/cobra"
)

func newCmdRun(ctx context.Context, opts RunOpts) *cobra.Command {
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

	cmd.AddGroup(&cobra.Group{
		ID:    "tasks",
		Title: "Available Tasks:",
	})

	if opts.Workflow == nil || opts.Workflow.HasErrors {
		return cmd
	}

	rootScope := &scope.Scope{
		Role:   scope.RoleRoot,
		Consts: opts.Workflow.File.Consts,
		Inputs: make(map[string]any),
		Globals: scope.Globals{
			Env: scope.Env(),
			Project: scope.NewProjectInfo(scope.ProjectInfoOpts{
				WorkDir:      opts.WorkDir,
				WorkspaceDir: opts.WorkDir,
				WorkflowFile: opts.Workflow.File.Path,
			}),
		},
	}

	err := addRootInputs(ctx, flagBindingOpts{
		logger: opts.Logger,
		cmd:    cmd,
		scope:  rootScope,
		inputs: opts.Workflow.File.Inputs,
	})

	if err != nil {
		opts.Logger.Error(err)
		return cmd
	}

	addTaskCommands(cmd, opts.Workflow.File)
	return cmd
}

type flagBindingOpts struct {
	logger *log.Logger
	scope  *scope.Scope
	cmd    *cobra.Command
	inputs manifest.Inputs
}

func (opts flagBindingOpts) inputBindingOpts() inputflag.InputBindingOpts {
	return inputflag.InputBindingOpts{
		EvalContext: scope.NewEvalContext(opts.scope),
		EnvVars:     opts.scope.Globals.Env,
		Scope:       opts.scope,
	}
}

func addRootInputs(ctx context.Context, opts flagBindingOpts) error {
	// TODO: add global inputs into a group
	binder := inputflag.NewInputFlagsBinder(ctx, opts.logger, opts.inputBindingOpts())

	for _, input := range opts.inputs {
		if err := binder.BindGlobalInput(input, opts.cmd); err != nil {
			return err
		}
	}

	return nil
}

func addTaskCommands(dst *cobra.Command, jf manifest.JobFile) {
	if len(jf.Tasks) == 0 {
		return
	}

	dst.Example = fmt.Sprintf("gilbert run %s")
	sampleTask := ""
	for name, task := range jf.Tasks {
		sampleTask = name
		shortDoc, longDoc := getTaskDescription(name, task)
		cmd := &cobra.Command{
			GroupID: "tasks",
			Use:     name + " [flags]",
			Short:   shortDoc,
			Long:    longDoc,
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd:   true,
				DisableNoDescFlag:   true,
				DisableDescriptions: true,
				HiddenDefaultCmd:    true,
			},
			RunE: func(cmd *cobra.Command, _ []string) error {
				cmd.Println("test!", name)
				return nil
			},
		}

		// TODO: mount flags
		dst.AddCommand(cmd)
	}

	dst.Example = fmt.Sprintf("$ gilbert run %s", sampleTask)
}

func getTaskDescription(name string, t *manifest.JobGroup) (string, string) {
	switch len(t.Doc) {
	case 0:
		str := fmt.Sprintf("Run %q task", name)
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
