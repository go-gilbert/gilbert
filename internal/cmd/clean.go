package cmd

import "github.com/spf13/cobra"

func newCmdClean(f Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean cached files and objects",
		Long:  "Cleans up cache and temporary files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
