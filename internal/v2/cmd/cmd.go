package cmd

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/signal"

	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/loader/yamlloader"
)

func Main(args []string) int {
	ctx, cancelFn := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFn()

	opts := cmdutil.BootstrapArgsFromFlags(args)

	logger := log.NewLogger("", opts.LogLevel, opts.BuildLogWriter())
	runOpts, err := buildRunOpts(ctx, logger, opts)
	if err != nil {
		// Log an error but still continue to build cli app.
		logger.Error(err)
	}

	cmd := newCmdRoot(runOpts)

	// Remove command name to avoid error when binary name doesn't match command.
	cmd.SetArgs(args[1:])
	err = cmd.ExecuteContext(ctx)
	exitCode := handleCmdError(logger, err)
	return exitCode
}

func handleCmdError(logger *log.Logger, err error) int {
	if err == nil || errors.Is(err, context.Canceled) {
		return 0
	}

	logger.Error(err)
	return 1
}

func buildRunOpts(ctx context.Context, logger *log.Logger, opts cmdutil.BootstrapOpts) (RunOpts, error) {
	runOpts := RunOpts{
		GlobalDefaults: opts,
		Logger:         logger,
	}

	workDir, err := opts.SetupWorkDir()
	if err != nil {
		return runOpts, err
	}

	// TODO: load workflow from cache if possible and opts.NoCache is false.
	runOpts.WorkDir = workDir
	runOpts.GlobalDefaults.WorkDir = workDir
	workflowFile, err := cmdutil.LocateWorkflowFile(workDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// bootstrap app w/o jobfile
			return runOpts, nil
		}

		return runOpts, err
	}

	// TODO: fill builtins list
	fileLoader := yamlloader.NewLoader(yamlloader.LoaderConfig{
		BuiltinNamespaces: nil,
	})

	runOpts.Workflow, err = fileLoader.Load(ctx, workflowFile)
	if err != nil {
		runOpts.WorkflowLoadError = err
	}

	return runOpts, err
}
