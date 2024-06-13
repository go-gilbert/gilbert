package tasks

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"
)

// FlagJSON sets output in JSON format
const FlagJSON = "json"

type tasksSummary struct {
	FileName string   `json:"file"`
	Tasks    []string `json:"tasks"`
}

// ListTasksAction handles 'ls' command
func ListTasksAction(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current working directory, %v", err)
	}

	m, err := manifest.FromDirectory(dir)
	if err != nil {
		return err
	}

	if ctx.Bool(FlagJSON) {
		// print tasks in JSON format if appropriate flag enabled
		return tasksToJSON(m)
	}

	if len(m.Tasks) == 0 {
		log.Default.Logf("No tasks defined in '%s'", m.Location())
		return nil
	}

	msg := fmt.Sprintf("List of defined tasks in '%s':", m.Location())
	for k := range m.Tasks {
		msg += "\n  - " + k
	}

	log.Default.Logf(msg)
	return nil
}

func tasksToJSON(m *manifest.Manifest) error {
	s := &tasksSummary{FileName: m.Location(), Tasks: make([]string, 0, len(m.Tasks))}
	for t := range m.Tasks {
		s.Tasks = append(s.Tasks, t)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
