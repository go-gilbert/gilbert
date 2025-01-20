package main

import (
	"os"

	"github.com/go-gilbert/gilbert/internal/cmd"
)

var (
	// These values will override by linker
	version = "dev"
	commit  = "local build"
)

func main() {
	app := cmd.NewCmdRoot(cmd.VersionInfo{
		Version: version,
		Commit:  commit,
	})

	err := app.Run(os.Args)
	cmd.Exit(err)
}
