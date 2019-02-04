package tasks

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/runner"
)

var (
	r *runner.TaskRunner
)

// RunTask is a handler for 'run' command
func RunTask(c *cli.Context) (err error) {
	if c.NArg() == 0 {
		return fmt.Errorf("no task specified")
	}

	task := c.Args()[0]

	r, err = getRunner()
	if err != nil {
		return err
	}

	return runTask(task)
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

	return runner.NewTaskRunner(m, dir, logging.Log), nil
}

func runTask(taskName string) error {
	if err := r.RunTask(taskName); err != nil {
		return err
	}

	logging.Log.Success("Task '%s' ran successfully\n", taskName)
	return nil
}
