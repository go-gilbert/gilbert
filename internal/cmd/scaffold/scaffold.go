package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/goccy/go-yaml"
	"github.com/urfave/cli"
)

var boilerplate = manifest.Manifest{
	Version: "1.0",
	Vars: manifest.Vars{
		"appVersion": "1.0.0",
	},
	Tasks: manifest.TaskSet{
		"build": manifest.Task{
			{
				Description: "Build project",
				ActionName:  "build",
			},
		},
		"cover": manifest.Task{
			{
				Description: "Check project coverage",
				ActionName:  "cover",
				Params: map[string]interface{}{
					"threshold":      60.0,
					"reportCoverage": true,
					"packages": []string{
						"./...",
					},
				},
			},
		},
		"clean": manifest.Task{
			{
				Description: "Remove vendor files",
				Condition:   "file ./vendor",
				ActionName:  "shell",
				Params: map[string]interface{}{
					"command": "rm -rf ./vendor",
				},
			},
		},
	},
}

// RunScaffoldManifest handles 'init' command
func RunScaffoldManifest(_ *cli.Context) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current working directory, %w", err)
	}

	log.Default.Debugf("current working directory is %q", cwd)
	out, err := yaml.Marshal(boilerplate)
	if err != nil {
		return fmt.Errorf("cannot create YAML file: %s", err)
	}

	destFile := filepath.Join(cwd, manifest.FileName)
	err = os.WriteFile(filepath.Join(cwd, manifest.FileName), out, 0655)
	if err != nil {
		return fmt.Errorf("failed to write '%q: %w", destFile, err)
	}

	log.Default.Successf("File %q successfully created", destFile)
	log.Default.Info("Use 'gilbert run build' to build the project")
	return nil
}
