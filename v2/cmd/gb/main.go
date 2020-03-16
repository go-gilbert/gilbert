package main

import (
	"fmt"
	"github.com/go-gilbert/gilbert/v2/cmd"
	"github.com/go-gilbert/gilbert/v2/manifest"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
)

const (
	fname    = "gilbert.hcl"
	Version  = "2.0.0-snapshot"
	CommitID = "dev"
)

var (
	verbose      = false
	disableColor = false

	exeName = filepath.Base(os.Args[0])

	rootCmd = &cobra.Command{
		Use:           "gb",
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Gilbert - a task runner for Go projects",
		Long: "Gilbert is task runner for Go projects\n\n" +
			"Complete documentation is available at https://go-gilbert.github.io",

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello World")
		},
	}

	lsCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   fmt.Sprintf("List tasks from %q file", manifest.DefaultFileName),
		Run:     cmd.WrapCobraCommand(listManifestCommand),
	}

	inspectCmd = &cobra.Command{
		Use:   "inspect [task name]",
		Long:  "Show task description and required parameters",
		Short: "Show task summary",
		Args:  cobra.ExactArgs(1),
		Run:   cmd.WrapCobraCommand(inspectManifestTask),
	}
)

func init() {
	fl := rootCmd.PersistentFlags()
	fl.BoolVarP(&verbose, "verbose", "v", false, "show debug information, useful for troubleshooting")
	fl.BoolVarP(&disableColor, "no-color", "n", false, "disable color output in terminal")
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf(
				"Gilbert version %s (%s) %s/%s\n\nGo version: %s\n",
				Version, CommitID, runtime.GOOS, runtime.GOARCH, runtime.Version(),
			)
		},
	})

	lsCmd.PersistentFlags().Bool("json", false, "Return data in JSON format")
	rootCmd.AddCommand(lsCmd)
}

func main() {
	cmd.ExitWithError(rootCmd.Execute())
	//
	//manPath, found, err := cmd.FindManifest2()
	//if err != nil {
	//	cmd.ExitWithError(err.Error())
	//}
	//
	//if found {
	//	m, err := cmd.LoadManifest(rootCmd, manPath)
	//	if err != nil {
	//		cmd.ExitWithError(err.Error())
	//	}
	//
	//	lsCmd.Run = cmd.PrintManifestCommandHandler(m, false)
	//}
	//
	//if err := rootCmd.Execute(); err != nil {
	//	_, _ = fmt.Fprintln(os.Stderr, err.Error())
	//	os.Exit(1)
	//}
}

func inspectManifestTask(c *cobra.Command, args []string) error {
	taskName := args[0]
	m, err := cmd.FindManifest()
	if err != nil {
		return err
	}

	t, ok := m.Tasks[taskName]
	if !ok {
		return fmt.Errorf("no such task %q", taskName)
	}

	fmt.Println("Name:\n", taskName)
	desc := "NOT AVAILABLE"
	if t.HasDescription() {
		desc = t.Description
	}

	fmt.Println("\nDescription:\n", desc)

	if t.HasParameters() {
		fmt.Println("\nParameters:")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Type", "Description"})

		for _, param := range t.Parameters {
			desc := param.Description
			if !param.IsRequired() {
				desc += " (required)"
			}
			table.Append([]string{param.Name, param.Type.FriendlyName(), desc})
		}

		table.Render()
	}
	return nil
}

func listManifestCommand(c *cobra.Command, args []string) error {
	m, err := cmd.FindManifest()
	if err != nil {
		return err
	}

	fmt.Printf("List of tasks defined in %q\n\n", m.FilePath())
	for _, task := range m.Tasks {
		fmt.Printf("- %q\t%s\n", task.Name, task.Description)
	}

	fmt.Printf("\nUse \"%s inspect\" to get information about task parameters.\n", exeName)
	return nil
}
