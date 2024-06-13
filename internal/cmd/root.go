package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-gilbert/gilbert/internal/config"
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	programName := filepath.Base(os.Args[0])

	cmd := &cobra.Command{
		Use:           fmt.Sprintf("%s <command> <subcommand> [flags]", programName),
		Short:         "Gilbert task runner",
		Long:          "Gilbert - Workflow automation tool",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       config.Version,
		Example: heredoc.Docf(`
			$ %[1]s run someTask --param1=foo --param2
			$ %[1]s ls
		`, programName),
	}

	versionOutput := fmt.Sprintf("{{.Name}} version {{.Version}} %s (%s)\n", config.Platform, config.Commit)
	cmd.SetVersionTemplate(versionOutput)

	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print program version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("%s version %s %s (%s)\n", programName, config.Version, config.Platform, config.Commit)
		},
	})
	return cmd
}
