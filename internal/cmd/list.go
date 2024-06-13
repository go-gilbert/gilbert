package cmd

import "github.com/spf13/cobra"

func newCmdList(f Factory) *cobra.Command {
	// TODO: add positional argument to filter generators and etc.
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all available tasks",
		Long:    "List all available tasks and schematics defined in Gilbert file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
