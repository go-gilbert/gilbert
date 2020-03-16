package cmd

import (
	"fmt"
	"github.com/go-gilbert/gilbert/support/fs"
	"github.com/go-gilbert/gilbert/v2/manifest"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type CommandHandler = func(c *cobra.Command, args []string) error

var BinName = filepath.Base(os.Args[0])

func FindManifest() (*manifest.Manifest, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %s", err)
	}

	manPath, found, err := fs.Lookup(manifest.DefaultFileName, wd, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to find file %q: %s", manifest.DefaultFileName, err.Error())
	}

	if !found {
		return nil, fmt.Errorf(
			`file %q not found in project directory. Use "%s init" to create a new one`,
			manifest.DefaultFileName, BinName,
		)
	}

	// TODO: prepare context
	return manifest.FromFile(manPath, nil)
}

func WrapCobraCommand(h CommandHandler) func(*cobra.Command, []string) {
	return func(c *cobra.Command, args []string) {
		ExitWithError(h(c, args))
	}
}

func ExitWithErrorMessage(msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	_, _ = fmt.Fprintln(os.Stderr, "error: ", msg)
	os.Exit(1)
}

func ExitWithError(err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintln(os.Stderr, "error: ", err.Error())
	os.Exit(1)
}
