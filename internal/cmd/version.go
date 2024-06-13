package cmd

import (
	"fmt"

	"github.com/go-gilbert/gilbert/internal/config"
	"github.com/spf13/cobra"
)

func newCmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print program version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf(
				"%s version %s %s (%s)\n",
				config.ProgramName, config.Version, config.Platform, config.Commit,
			)
		},
	}
}
