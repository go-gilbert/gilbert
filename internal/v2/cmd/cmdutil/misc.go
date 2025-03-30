package cmdutil

import "github.com/spf13/cobra"

func DecorateHelpFunc(cmd *cobra.Command, fn func()) {
	origHelp := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fn()
		origHelp(cmd, args)
	})
}
