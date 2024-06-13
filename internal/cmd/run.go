package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-gilbert/gilbert/internal/config"
	"github.com/spf13/cobra"
)

func newCmdRun(f Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run <task> [parameters]",
		Aliases: []string{"r"},
		Short:   "Run a task",

		// TODO: add doc URL.
		Long: heredoc.Doc(`
			Runs a task declared in Gilbert file (gilbert.yaml)
		`),
		Example: heredoc.Docf(`
			$ %[1]s run some-task --param1 foo --param2=bar 
		`, config.ProgramName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
