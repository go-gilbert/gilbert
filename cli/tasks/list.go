package tasks

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	"github.com/x1unix/gilbert/logging"
)

// ListTasksAction handles 'ls' command
func ListTasksAction(_ *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current working directory, %v", err)
	}

	m, err := getManifest(dir)
	if err != nil {
		return err
	}

	if len(m.Tasks) == 0 {
		log.Log.Log("No tasks defined in '%s'", m.Location())
		return nil
	}

	msg := fmt.Sprintf("List of defined tasks in '%s':", m.Location())
	for k := range m.Tasks {
		msg += "\n  - " + k
	}

	log.Log.Log(msg)
	return nil
}
