package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/go-gilbert/gilbert/v2/internal/config"
	"github.com/go-gilbert/gilbert/v2/internal/log"
	"github.com/go-gilbert/gilbert/v2/internal/spec"
	"github.com/spf13/cobra"
)

var (
	// AppVersion is application version in semver format. Provided by linker.
	AppVersion = "v0.0.1"

	// ReleaseType is application release type (snapshot, beta, stable). Provided by linker.
	ReleaseType = "snapshot"

	// CommitSHA is git commit SHA. Provided by linker.
	CommitSHA = "000000"
)

func Run(ctx context.Context, cfg *config.CoreConfig, args []string) {
	logger, err := cfg.NewLogger()
	if err != nil {
		log.Global().Fatal(err)
	}

	if err := RunE(ctx, logger, cfg, args); err != nil {
		logger.Fatal(err)
	}
}

func RunE(ctx context.Context, logger log.Printer, cfg *config.CoreConfig, args []string) error {
	ctx, cancelFn := NewApplicationContext(ctx)
	defer cancelFn()

	cmd := NewCmdRoot(cfg)
	cmd.SetArgs(args)

	projectSpec := cfg.ProjectSpec()
	data, err := os.ReadFile(projectSpec.FileName)
	if os.IsNotExist(err) {
		// Ignore error
		logger.Debugf("file not found: %q, skipping spec read", projectSpec.FileName)
		return cmd.ExecuteContext(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	parser := spec.NewParser(spec.NewRootContext(), projectSpec)
	_, err = parser.Parse(data)
	if err != nil {
		var diags hcl.Diagnostics
		if errors.As(err, &diags) {
			logger.ReportDiagnostics(data, diags)
		}

		return err
	}

	err = cmd.ExecuteContext(ctx)
	return err
}

func NewCmdRoot(coreCfg *config.CoreConfig) *cobra.Command {
	cfg := &config.Config{
		CoreConfig: *coreCfg,
	}

	cmd := &cobra.Command{
		Use:           "gilbert <command> <subcommand> [flags]",
		Short:         "Gilbert task runner",
		Long:          "gilbert - Gilbert task runner command line tool",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       AppVersion,
		Example: heredoc.Doc(`
			$ gilbert ls
			$ gilbert run ...
			$ gilbert generate ...
		`),
	}

	versionOutput := fmt.Sprintln("gilbert version", AppVersion)
	cmd.AddCommand(&cobra.Command{
		Use:    "version",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(versionOutput)
		},
	})

	cmd.SetArgs(ExpandedArgs())
	cmd.PersistentFlags().StringVar(&cfg.CacheDir, "cache-dir", ".gilbert/cache",
		"specify a custom folder that must be used to store cache")
	cmd.PersistentFlags().StringVar(&cfg.WorkDir, flagCwd, coreCfg.WorkDir, "working directory to use")
	cmd.PersistentFlags().StringVar(&cfg.SpecFile, flagSpecFile, coreCfg.SpecFile, "use other file as Gilbert file")
	cmd.PersistentFlags().StringVar(&cfg.LogFormat, flagLogFormat, coreCfg.LogFormat, "set console output format (color,text,json)")
	cmd.PersistentFlags().BoolVar(&cfg.Verbose, flagVerbose, coreCfg.Verbose, "print debugging information")
	cmd.Flags().Bool("version", false, "Show hubctl version")
	cmd.PersistentFlags().Bool("help", false, "Show help for command")
	return cmd
}
