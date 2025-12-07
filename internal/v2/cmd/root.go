package cmd

import (
	"context"
	"fmt"
	"runtime"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/go-gilbert/gilbert/internal/v2/cmd/cmdutil"
	"github.com/go-gilbert/gilbert/internal/v2/cmd/help"
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/loader/yamlloader"
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
		Version:       fmt.Sprintf("%s %s/%s", "snapshot", runtime.GOOS, runtime.GOARCH),
		SilenceErrors: true,
		SilenceUsage:  true,
		Example: heredoc.Doc(`
			$ gilbert init
			$ gilbert list
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

	cmd.SetUsageFunc(help.NewGeneralUsageFunc(cmdutil.NewUsageColorPalette(opts.GlobalDefaults.NoColor)))

	cmd.PersistentFlags().Bool("help", false, "Show help for command")
	fset := buildGlobalsFlagSet(&opts.GlobalDefaults)
	cmd.PersistentFlags().AddFlagSet(fset)

	cmd.AddCommand(newCmdRun(ctx, opts, fset))
	cmd.AddCommand(newCmdDiagnostics(opts))
	cmd.AddCommand(newCmdList(opts))
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

// buildGlobalsFlagSet constructs and returns flag set with global flags.
//
// Although core flags were already processed by uflag before, it's still useful to validate flag values and
// display then in help output.
func buildGlobalsFlagSet(opts *cmdutil.BootstrapOpts) *pflag.FlagSet {
	fset := pflag.NewFlagSet("globals", pflag.ExitOnError)
	defaultLogFormat := log.FormatConsole
	switch true {
	case opts.JSON:
		defaultLogFormat = log.FormatJSON
	case opts.NoColor:
		defaultLogFormat = log.FormatNoColor
	}

	// Log writer was already set by uflag, just add flag for docs.
	fset.Var(log.NewFlagFormat(nil, defaultLogFormat), cmdutil.FlagLogFormat, "Set output log format")
	fset.Var(log.NewFlagLevel(&opts.LogLevel), cmdutil.FlagLogLevel, "Set output log level")
	fset.StringVar(&opts.WorkDir, cmdutil.FlagWorkDir, opts.WorkDir, "Working directory to use")
	fset.BoolVar(&opts.NoCache, cmdutil.FlagNoCache, opts.NoCache, "Disable caches")
	return fset
}
