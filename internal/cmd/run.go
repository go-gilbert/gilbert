package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/go-gilbert/gilbert/internal/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/cmd/cmdutil/inputflag"
	"github.com/go-gilbert/gilbert/internal/cmd/help"
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type runContext struct {
	jobFile   manifest.JobFile
	rootScope *scope.Scope
}

// newCmdRun constructs a run command with mounted flags and sub-commands from a workflow file.
//
// globalFlags passed will be used for per-task usage help func.
func newCmdRun(ctx context.Context, opts RunOpts, globalFlags *pflag.FlagSet) *cobra.Command {
	diagRenderer := cmdutil.NewDiagnosticsRenderer(opts.Logger, opts.GlobalDefaults)

	var runCtx runContext
	mp := taskFlagMountParams{
		logger:      opts.Logger,
		defaults:    opts.GlobalDefaults,
		globalFlags: globalFlags,
		inputDiags:  inputflag.NewDiagnosticsCollector(),
	}

	if opts.Workflow != nil {
		mp.fileDiags = opts.Workflow.Diagnostics
		runCtx.jobFile = opts.Workflow.File
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// PreRunE is executed before cobra performs flag validation checks.
			// Use this to test basic sanity checks.
			if len(args) == 0 {
				// Disallow calling "run" without task name.
				return errTaskNameRequired
			}

			return handleTaskNotFound(opts, args)
		},
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if opts.Workflow == nil {
				return nil
			}

			diagRenderer.RenderDiagnostics(opts.Workflow.Diagnostics)
			diagRenderer.RenderDiagnostics(mp.inputDiags.Diagnostics)
			diagRenderer.Reset()

			if opts.Workflow.HasErrors {
				return errors.New("workflow file contains errors")
			}

			if mp.inputDiags.HasErrors {
				return errors.New("error occurred when reading workflow inputs")
			}

			return nil
		},
	}

	// Render inputs diagnostics in help to indicate why some flags or defaults are missing.
	cmdutil.DecorateHelpFunc(cmd, func() {
		diagRenderer.RenderDiagnostics(mp.inputDiags.Diagnostics)
	})

	// Cobra flags won't be mounted if workflow file has errors.
	// If user calls "run" command with broken workflow - Cobra just throws "unknown flag" error.
	// To avoid user confusion - render file diagnostics before exit.
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		if opts.Workflow == nil {
			return err
		}

		if len(opts.Workflow.Diagnostics) > 0 {
			diagRenderer.RenderDiagnostics(opts.Workflow.Diagnostics)
		}

		// swallow error to avoid user's confusion.
		if opts.Workflow.HasErrors {
			return errors.New("workflow file contains errors")
		}

		return err
	})

	runHelp := &help.RunUsageInfo{
		NoColor:     opts.GlobalDefaults.NoColor,
		GlobalFlags: globalFlags,
	}

	cmd.SetUsageFunc(help.NewRunUsageFunc(runHelp))
	cmd.AddGroup(&cobra.Group{
		ID:    "tasks",
		Title: "Available Tasks:",
	})

	if opts.Workflow == nil || opts.Workflow.HasErrors {
		return cmd
	}

	runCtx.rootScope = &scope.Scope{
		Consts: opts.Workflow.File.Consts,
		Inputs: make(map[string]any),
		Globals: scope.Globals{
			Env: scope.Env(),
			Project: scope.ProjectInfo{
				WorkDir:      opts.WorkDir,
				WorkspaceDir: opts.WorkDir,
				WorkflowFile: opts.Workflow.File.Path,
			},
		},
	}

	workflowFlags, err := addRootInputs(ctx, cmd, flagBindingOpts{
		logger: opts.Logger,
		scope:  runCtx.rootScope,
		inputs: opts.Workflow.File.Inputs,
		diags:  mp.inputDiags,
	})
	runHelp.WorkflowFlags = workflowFlags

	if err != nil {
		opts.Logger.Error(err)
		return cmd
	}

	mp.workflowFlags = workflowFlags
	if err := addTaskCommands(ctx, cmd, mp, runCtx); err != nil {
		opts.Logger.Error(err)
	}

	return cmd
}

type flagBindingOpts struct {
	logger *log.Logger
	scope  *scope.Scope
	inputs manifest.Inputs
	diags  *inputflag.DiagnosticsCollector
}

func (opts flagBindingOpts) inputBindingOpts() inputflag.InputBindingOpts {
	return inputflag.InputBindingOpts{
		EvalParams:  scope.NewEvalParams(opts.scope),
		EnvVars:     opts.scope.Globals.Env,
		Scope:       opts.scope,
		Diagnostics: opts.diags,
	}
}

func addRootInputs(ctx context.Context, cmd *cobra.Command, opts flagBindingOpts) (*pflag.FlagSet, error) {
	binder := inputflag.NewInputFlagsBinder(opts.logger, opts.inputBindingOpts())

	fset := pflag.NewFlagSet("workflow", pflag.ExitOnError)
	requiredFields, err := binder.BindInputsToFlagSet(ctx, fset, true, opts.inputs)
	if err != nil {
		return nil, err
	}

	cmd.PersistentFlags().AddFlagSet(fset)
	for _, fname := range requiredFields {
		if err := cmd.MarkPersistentFlagRequired(fname); err != nil {
			return nil, fmt.Errorf("can't mark flag %q as required: %w", fname, err)
		}
	}

	return fset, nil
}

type taskFlagMountParams struct {
	logger        *log.Logger
	defaults      cmdutil.BootstrapOpts
	inputDiags    *inputflag.DiagnosticsCollector
	fileDiags     parsetypes.Diagnostics
	globalFlags   *pflag.FlagSet
	workflowFlags *pflag.FlagSet
}

func addTaskCommands(ctx context.Context, dst *cobra.Command, fp taskFlagMountParams, runCtx runContext) error {
	if len(runCtx.jobFile.Tasks) == 0 {
		return nil
	}

	sampleTask := ""
	for name, task := range runCtx.jobFile.Tasks {
		sampleTask = name
		shortDoc, longDoc := getTaskDescription(name, task)
		taskScope := runCtx.rootScope.Fork()
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
				return startTaskRunner(cmd, taskRunConfig{
					taskName:     name,
					logger:       fp.logger,
					jobFile:      &runCtx.jobFile,
					scope:        taskScope,
					bootstapOpts: fp.defaults,
				})
			},
		}

		fset := pflag.NewFlagSet("task", pflag.ExitOnError)
		if len(task.Inputs) > 0 {
			bindOpts := flagBindingOpts{
				logger: fp.logger,
				scope:  taskScope,
				inputs: task.Inputs,
				diags:  &inputflag.DiagnosticsCollector{},
			}

			binder := inputflag.NewInputFlagsBinder(bindOpts.logger, bindOpts.inputBindingOpts())
			requiredFlags, err := binder.BindInputsToFlagSet(ctx, fset, false, bindOpts.inputs)
			if err != nil {
				return err
			}

			cmd.Flags().AddFlagSet(fset)
			for _, fname := range requiredFlags {
				if err := cmd.MarkFlagRequired(fname); err != nil {
					return fmt.Errorf("can't mark flag %q of task %q as required: %w", fname, name, err)
				}
			}
		}

		// TODO: gen usage
		usageFunc := help.NewTaskUsageFunc(help.TaskUsageInfo{
			NoColor:       fp.defaults.NoColor,
			GlobalFlags:   fp.globalFlags,
			WorkflowFlags: fp.workflowFlags,
			TaskFlags:     fset,
		})

		cmd.SetUsageFunc(usageFunc)
		dst.AddCommand(cmd)
	}

	dst.Example = fmt.Sprintf("$ gilbert run %s", sampleTask)
	return nil
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
		return errTaskNameRequired
	}

	if opts.WorkflowLoadError != nil {
		return fmt.Errorf("workflow file cannot be loaded: %w", opts.WorkflowLoadError)
	}

	if opts.Workflow == nil {
		return errWorkflowNotFound
	}

	if opts.Workflow.HasErrors {
		return newErrWorkflowFileHasErrors(opts.Workflow.File.Path)
	}

	return newErrTaskNotFound(args[0])
}
