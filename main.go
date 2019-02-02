package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli"
	"github.com/x1unix/gilbert/cli/scaffold"
	"github.com/x1unix/gilbert/cli/tasks"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
)

var (
	Version = "dev"
	Commit  = "null"
)

// unfortunately, urface/cli ignores '--verbose' global flag :(
var verboseFlag = cli.BoolFlag{
	Name:        "verbose",
	Usage:       "shows debug information, useful for troubleshooting",
	Destination: &scope.Debug,
}

func main() {
	app := cli.NewApp()
	app.Name = "gilbert"
	app.Usage = "Build automation tool for Go"
	app.Version = Version
	app.HideVersion = true
	app.Commands = []cli.Command{
		{
			Name:        "version",
			Description: "shows application version",
			Usage:       "Shows application version",
			Action:      printVersion,
		},
		{
			Name:        "run",
			Description: "Runs a task declared in manifest file",
			Usage:       "Runs a task declared in manifest file",
			Action:      tasks.RunTask,
			Before:      bootstrap,
			Flags: []cli.Flag{
				verboseFlag,
			},
		},
		{
			Name:        "init",
			Description: "Scaffolds a new gilbert.yaml file",
			Usage:       "Scaffolds a new gilbert.yaml file",
			Action:      scaffold.ScaffoldManifest,
			Before:      bootstrap,
			Flags: []cli.Flag{
				verboseFlag,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		color.Red("ERROR: %v", err)
	}
}

func bootstrap(_ *cli.Context) error {
	logging.Log = logging.NewConsoleLogger(logging.DefaultPadding, scope.Debug)
	return nil
}

func printVersion(_ *cli.Context) error {
	fmt.Printf("Gilbert version %s (%s)\n", Version, Commit)
	return nil
}
