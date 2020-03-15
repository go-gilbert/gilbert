package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-gilbert/gilbert/support/fs"

	"github.com/spf13/cobra"

	"github.com/go-gilbert/gilbert/v2/manifest"
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
		Use:   "gb",
		Short: "Gilbert - a task runner for Go projects",
		Long: "Gilbert is task runner for Go projects\n\n" +
			"Complete documentation is available at https://go-gilbert.github.io",

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello World")
		},
	}

	lsCmd = &cobra.Command{
		Use:   "ls",
		Short: fmt.Sprintf("List tasks from %q file", manifest.DefaultFileName),
		Run: func(cmd *cobra.Command, args []string) {
			// Stub for "gb ls" command if no manifest was found
			exitWithError(
				"%q not found. Run \"%s init\" to create a new one.",
				manifest.DefaultFileName, exeName,
			)
		},
	}
)

func init() {
	fl := rootCmd.PersistentFlags()
	fl.BoolVarP(&verbose, "verbose", "v", false, "show debug information, useful for troubleshooting")
	fl.BoolVarP(&disableColor, "no-color", "n", false, "disable color output in terminal")
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
	_, _, err := findManifest()
	if err != nil {
		exitWithError(err.Error())
	}

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func findManifest() (string, bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("failed to get working directory: %s", err)
	}

	manPath, found, err := fs.Lookup(manifest.DefaultFileName, wd, 3)
	if err != nil {
		return "", false, fmt.Errorf("failed to find file %q: %s", manifest.DefaultFileName, err.Error())
	}

	return manPath, found, nil
}

func exitWithError(msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	_, _ = fmt.Fprintln(os.Stderr, "error: ", msg)
	os.Exit(1)
}

func main2() {
	man, err := manifest.FromFile(fname, nil)
	if err != nil {
		switch t := err.(type) {
		case *manifest.Error:
			fmt.Println(t.PrettyPrint())
		default:
			fmt.Println(err)
		}
		os.Exit(1)
	}

	fmt.Println(man)
}

func must(err error) {
	if err == nil {
		return
	}

	panic(err)
}
