package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-gilbert/gilbert/internal/config"
	"github.com/spf13/cobra"
)

func newCmdInit(f Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [template]",
		Short: "Scaffold a new project",
		Long: heredoc.Doc(`
			Scaffolds a new gilbert file with optional project template.

			Template path is Git HTTP or SSH URL of a template source with an optional tag or git revision to use.
			Without template path, generates a simple gilbert.yaml file with sample project config.
		`),
		Args: cobra.MaximumNArgs(1),
		Example: heredoc.Docf(`
			$ %[1]s init
			$ %[1]s init https://github.com/username/some-template@v1.2.3
			$ %[1]s init git@gitlab.com/username/template-repo.git
		`, config.ProgramName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
