package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/urfave/cli"
	"github.com/x1unix/guru/logging"
	"github.com/x1unix/guru/manifest"
	"github.com/x1unix/guru/runner"
	"github.com/x1unix/guru/scope"
)

var (
	Version   = "0.0.1-alpha"
	r         *runner.TaskRunner
	subLogger logging.Logger
)

func main() {
	app := cli.NewApp()
	app.Name = "guru"
	app.Usage = "Build automation tool for Go"
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "shows debug information, useful for troubleshooting",
		},
	}

	app.Commands = []cli.Command{
		{
			Name: "version",
			Action: func(c *cli.Context) error {
				version()
				return nil
			},
		},
		{
			Name:        "run",
			Description: "Runs a task declared in manifest file",
			Action:      evalTask,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "verbose",
					Usage: "shows debug information, useful for troubleshooting",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		color.Red("ERROR: %v", err)
	}
}

func evalTask(c *cli.Context) (err error) {
	if c.Bool("verbose") {
		scope.Debug = true
	}

	if c.NArg() == 0 {
		return fmt.Errorf("no task specified")
	}

	task := c.Args()[0]
	logging.Log = logging.NewConsoleLogger(logging.DefaultPadding, scope.Debug)
	subLogger = logging.Log.SubLogger()

	r, err = getRunner()
	if err != nil {
		return err
	}

	err = runTask(task, c.Args())
	return
}

func version() {
	fmt.Println(Version)
}

func getRunner() (*runner.TaskRunner, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot get current working directory, %v", err)
	}

	data, err := ioutil.ReadFile(filepath.Join(dir, manifest.FileName))
	if err != nil {
		return nil, fmt.Errorf("manifest file not found (%s) at %s", manifest.FileName, dir)
	}

	m, err := manifest.UnmarshalManifest(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file:\n  %v", err)
	}

	return runner.NewTaskRunner(m, dir, subLogger), nil
}

func runTask(taskName string, args cli.Args) error {
	task, ok := r.TaskByName(taskName)
	if !ok {
		return fmt.Errorf("task '%s' doesn't exists", taskName)
	}

	logging.Log.Log("Running task '%s'...", taskName)
	steps := len(*task)

	for jobIndex, job := range *task {
		currentStep := jobIndex + 1
		descr := ""
		if job.HasDescription() {
			descr = ": " + job.Description
		}

		subLogger.Log("Step %d of %d%s", currentStep, steps, descr)
		err := r.RunJob(&job)
		if err != nil {
			return fmt.Errorf("task '%s' returned an error on step %d: %v", taskName, currentStep, err)
		}
	}

	logging.Log.Success("Task '%s' ran successfully", taskName)
	return nil
}
