package tasks

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/runner"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	subLogger logging.Logger
	r         *runner.TaskRunner
)

// RunTask is a handler for 'run' command
func RunTask(c *cli.Context) (err error) {
	if c.NArg() == 0 {
		return fmt.Errorf("no task specified")
	}

	task := c.Args()[0]
	subLogger = logging.Log.SubLogger()

	r, err = getRunner()
	if err != nil {
		return err
	}

	return runTask(task, c.Args())
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

	logging.Log.Success("Task '%s' ran successfully\n", taskName)
	return nil
}
