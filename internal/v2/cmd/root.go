package cmd

import (
	"context"
	"runtime"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/loader/yamlloader"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RunOpts struct {
	GlobalDefaults    cmdutil.BootstrapOpts
	WorkDir           string
	Logger            *log.Logger
	Workflow          *yamlloader.LoadResult
	WorkflowLoadError error
}

func newCmdRoot(ctx context.Context, opts RunOpts) *cobra.Command {
	printDebugRunOpts(opts)
	cmd := &cobra.Command{
		Use:           "gilbert <command> <subcommand> [flags]",
		Short:         "Gilbert task runner",
		SilenceErrors: true,
		SilenceUsage:  true,
		Example: heredoc.Doc(`
			$ gilbert init
			$ gilbert run foobar
		`),
		Annotations: map[string]string{
			// TODO: fill version from build info.
			"versionInfo": "snapshot",
			"platform":    runtime.GOOS + "/" + runtime.GOARCH,
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			opts.Logger.Infof("%#v", opts)
			return nil
		},
	}

	cmd.AddGroup(&cobra.Group{
		ID:    "tasks",
		Title: "Available tasks",
	})

	cmd.PersistentFlags().Bool("help", false, "Show help for command")
	mountCoreGlobalFlags(&opts.GlobalDefaults, cmd.PersistentFlags())
	if opts.Workflow != nil && !opts.Workflow.HasErrors {
		mountWorkflowInputsFlags(opts.Workflow.File.Inputs, cmd.PersistentFlags())
	}

	cmd.AddCommand(newCmdRun(ctx, opts))
	return cmd
}

func printDebugRunOpts(opts RunOpts) {
	logger := opts.Logger
	logger.Debugf("working directory: %q", opts.WorkDir)
	if opts.Workflow == nil {
		logger.Debug("workflow file not available")
		return
	}

	logger.Debugf("using workflow file: %q", opts.Workflow.File.Path)
	if opts.Workflow.HasErrors {
		logger.Warnf("workflow file %q has errors", opts.Workflow.File.Path)
	}
}

func mountWorkflowInputsFlags(inputs manifest.Inputs, flagSet *pflag.FlagSet) {
	// TODO: implement
}

// mountCoreGlobalFlags mounts core options as cobra command flags only for documentation purposes.
//
// Although core flags were already processed by uflag before, it's still useful to validate flag values and
// display then in help output.
func mountCoreGlobalFlags(opts *cmdutil.BootstrapOpts, flagSet *pflag.FlagSet) {
	defaultLogFormat := log.FormatConsole
	switch true {
	case opts.JSON:
		defaultLogFormat = log.FormatJSON
	case opts.NoColor:
		defaultLogFormat = log.FormatNoColor
	}

	// Log writer was already set by uflag, just add flag for docs.
	flagSet.Var(log.NewFlagFormat(nil, defaultLogFormat), cmdutil.FlagLogFormat, "Set output log format")
	flagSet.Var(log.NewFlagLevel(&opts.LogLevel), cmdutil.FlagLogLevel, "Set output log level")
	flagSet.StringVar(&opts.WorkDir, cmdutil.FlagWorkDir, opts.WorkDir, "Working directory to use")
	flagSet.BoolVar(&opts.NoCache, cmdutil.FlagNoCache, opts.NoCache, "Disable caches")
}
