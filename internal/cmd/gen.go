package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-gilbert/gilbert/internal/config"
	"github.com/spf13/cobra"
)

func newCmdGen(f Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gen <schematic> [parameters]",
		Aliases: []string{"g"},
		Short:   "Generates and/or modifies files based on a schematic.",

		// TODO: add docs
		Long: heredoc.Doc(`
			Generates and/or modifies files or directories based on a schematic.
	
			List of available schematics is provided by third-party plugins and schematic definitions in a Gilbert file.
		`),
		Example: heredoc.Docf(`
			$ %[1]s new go.test somepackage/file_test.go --testify
		`, config.ProgramName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
