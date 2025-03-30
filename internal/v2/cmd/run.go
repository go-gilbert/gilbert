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
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/spf13/cobra"
)

type mountContext struct {
	logger     *log.Logger
	defaults   cmdutil.BootstrapOpts
	inputDiags *inputflag.DiagnosticsCollector
	fileDiags  parsetypes.Diagnostics
}

func newCmdRun(ctx context.Context, opts RunOpts) *cobra.Command {
	mountCtx := mountContext{
		logger:     opts.Logger,
		defaults:   opts.GlobalDefaults,
		inputDiags: inputflag.NewDiagnosticsCollector(),
	}

	if opts.Workflow != nil {
		mountCtx.fileDiags = opts.Workflow.Diagnostics
	}

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
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if opts.Workflow == nil {
				return nil
			}

			cmdutil.RenderDiagnostics(opts.Logger, opts.GlobalDefaults, opts.Workflow.Diagnostics)
			cmdutil.RenderDiagnostics(opts.Logger, opts.GlobalDefaults, mountCtx.inputDiags.Diagnostics)

			if opts.Workflow.HasErrors {
				return errors.New("workflow file contains errors")
			}

			if mountCtx.inputDiags.HasErrors {
				return errors.New("error occurred when reading workflow inputs")
			}

			return nil
		},
	}

	// Render inputs diagnostics in help to indicate why some flags or defaults are missing.
	cmdutil.DecorateHelpFunc(cmd, func() {
		cmdutil.RenderDiagnostics(opts.Logger, opts.GlobalDefaults, mountCtx.inputDiags.Diagnostics)
	})

	// Cobra flags won't be mounted if workflow file has errors.
	// If user calls "run" command with broken workflow - Cobra just throws "unknown flag" error.
	// To avoid user confusion - render file diagnostics before exit.
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		if opts.Workflow == nil {
			return err
		}

		if len(opts.Workflow.Diagnostics) > 0 {
			cmdutil.RenderDiagnostics(opts.Logger, opts.GlobalDefaults, opts.Workflow.Diagnostics)
		}

		// swallow error to avoid user's confusion.
		if opts.Workflow.HasErrors {
			return errors.New("workflow file contains errors")
		}

		return err
	})

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
		diags:  mountCtx.inputDiags,
	})

	if err != nil {
		opts.Logger.Error(err)
		return cmd
	}

	addTaskCommands(cmd, opts.Workflow.File, mountCtx)
	return cmd
}

type flagBindingOpts struct {
	logger *log.Logger
	scope  *scope.Scope
	cmd    *cobra.Command
	inputs manifest.Inputs
	diags  *inputflag.DiagnosticsCollector
}

func (opts flagBindingOpts) inputBindingOpts() inputflag.InputBindingOpts {
	return inputflag.InputBindingOpts{
		EvalContext: scope.NewEvalContext(opts.scope),
		EnvVars:     opts.scope.Globals.Env,
		Scope:       opts.scope,
		Diagnostics: opts.diags,
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

func addTaskCommands(dst *cobra.Command, jf manifest.JobFile, mctx mountContext) {
	if len(jf.Tasks) == 0 {
		return
	}

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
