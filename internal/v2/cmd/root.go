package cmd

import (
	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/loader/yamlloader"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RunOpts struct {
	GlobalDefaults BootstrapOpts
	WorkDir        string
	Logger         *log.Logger
	Workflow       *yamlloader.LoadResult
}

func newCmdRoot(opts RunOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use: "gilbert",
		RunE: func(_ *cobra.Command, _ []string) error {
			opts.Logger.Infof("%#v", opts)
			return nil
		},
	}

	mountCoreGlobalFlags(&opts.GlobalDefaults, cmd.PersistentFlags())
	return cmd
}

func mountWorkflowGlobalFlags(f manifest.JobFile, flagSet *pflag.FlagSet) {
	
}

// mountCoreGlobalFlags mounts core options as cobra command flags only for documentation purposes.
//
// Although core flags were already processed by uflag before, it's still useful to validate flag values and
// display then in help output.
func mountCoreGlobalFlags(opts *BootstrapOpts, flagSet *pflag.FlagSet) {
	defaultLogFormat := log.FormatConsole
	switch true {
	case opts.JSON:
		defaultLogFormat = log.FormatJSON
	case opts.NoColor:
		defaultLogFormat = log.FormatNoColor
	}

	// Log writer was already set by uflag, just add flag for docs.
	flagSet.Var(log.NewFlagFormat(nil, defaultLogFormat), flagLogFormat, "set output log format")
	flagSet.Var(log.NewFlagLevel(&opts.LogLevel), flagLogLevel, "set output log level")
	flagSet.StringVar(&opts.WorkDir, flagWorkDir, opts.WorkDir, "working directory to use")
	flagSet.BoolVar(&opts.NoCache, flagNoCache, opts.NoCache, "disable caches")
}
