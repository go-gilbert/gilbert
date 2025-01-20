package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/go-gilbert/gilbert/internal/cmd/maintenance"
	"github.com/go-gilbert/gilbert/internal/cmd/scaffold"
	"github.com/go-gilbert/gilbert/internal/cmd/tasks"
	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/urfave/cli"
)

// FlagNoColor disables color output
const FlagNoColor = "no-color"

var (
	// unfortunately, urfave/cmd ignores '--verbose' global flag :(
	// so it should be defined implicitly in each task
	verboseFlag = cli.BoolFlag{
		Name:        "verbose",
		Usage:       "shows debug information, useful for troubleshooting",
		Destination: &scope.Debug,
	}

	noColorFlag = cli.BoolFlag{
		Name:  FlagNoColor,
		Usage: "disable color output in terminal",
	}
)

type VersionInfo struct {
	Version string
	Commit  string
}

func NewCmdRoot(ver VersionInfo) *cli.App {
	app := cli.NewApp()
	app.Name = "gilbert"
	app.Usage = "Build automation tool for Go"
	app.Version = ver.Version
	app.HideVersion = true
	app.Commands = []cli.Command{
		{
			Name:        "version",
			Description: "shows application version",
			Usage:       "Shows application version",
			Action: func(_ *cli.Context) error {
				fmt.Printf("Gilbert version %s (%s)\n", ver.Version, ver.Commit)
				return nil
			},
		},
		{
			Name:        "run",
			Description: "Runs a task declared in manifest file",
			Usage:       "Runs a task declared in manifest file",
			Action:      tasks.RunTask,
			Before:      bootstrap,
			Flags: []cli.Flag{
				verboseFlag,
				noColorFlag,
				cli.StringSliceFlag{
					Name: tasks.OverrideVarFlag,
				},
			},
		},
		{
			Name:        "ls",
			Description: "Lists all tasks defiled in gilbert.yaml",
			Usage:       "Lists all tasks defiled in gilbert.yaml",
			Action:      tasks.ListTasksAction,
			Before:      bootstrap,
			Flags: []cli.Flag{
				verboseFlag,
				cli.BoolFlag{
					Name:  tasks.FlagJSON,
					Usage: "Print output in JSON format",
				},
			},
		},
		{
			Name:        "init",
			Description: "Scaffolds a new gilbert.yaml file",
			Usage:       "Scaffolds a new gilbert.yaml file",
			Action:      scaffold.RunScaffoldManifest,
			Before:      bootstrap,
			Flags: []cli.Flag{
				verboseFlag,
			},
		},
		{
			Name:        "clean",
			Description: "Clean cached files and objects",
			Usage:       "Clean cached files and objects",
			Action:      maintenance.ClearCacheAction,
			Before:      bootstrap,
			Flags: []cli.Flag{
				verboseFlag,
				maintenance.ClearAllFlag,
				maintenance.ClearPluginsFlag,
			},
		},
	}

	return app
}

func Exit(err error) {
	if err != nil {
		color.Red("ERROR: %v", err)
		os.Exit(1)
	}
}

func bootstrap(c *cli.Context) error {
	level := log.LevelInfo
	if scope.Debug {
		level = log.LevelDebug
	}

	noColor := c.Bool(FlagNoColor)
	log.UseConsoleLogger(level, noColor)
	return nil
}
