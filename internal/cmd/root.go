package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-gilbert/gilbert/internal/config"
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	cfg := &config.LaunchParams{}
	f := NewFactory(cfg)

	cmd := &cobra.Command{
		Use:           fmt.Sprintf("%s <command> <subcommand> [flags]", config.ProgramName),
		Short:         "Gilbert task runner",
		Long:          "Gilbert - Workflow automation tool",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       config.Version,
		Example: heredoc.Docf(`
			$ %[1]s run someTask --param1=foo --param2
			$ %[1]s ls
		`, config.ProgramName),
	}

	versionOutput := fmt.Sprintf("{{.Name}} version {{.Version}} %s (%s)\n", config.Platform, config.Commit)
	cmd.SetVersionTemplate(versionOutput)

	cmd.AddCommand(
		newCmdRun(f),
		newCmdInit(f),
		newCmdGen(f),
		newCmdList(f),
		newCmdClean(f),
		newCmdVersion(),
	)

	return cmd
}
