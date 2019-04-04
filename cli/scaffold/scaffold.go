package scaffold

import (
	"fmt"
	"github.com/x1unix/gilbert/log"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/urfave/cli"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
	"gopkg.in/yaml.v2"
)

var boilerplate = manifest.Manifest{
	Version: "1.0",
	Vars: scope.Vars{
		"appVersion": "1.0.0",
	},
	Tasks: manifest.TaskSet{
		"build": manifest.Task{
			{
				Description: "Build project",
				PluginName:  "build",
			},
		},
		"cover": manifest.Task{
			{
				Description: "Check project coverage",
				PluginName:  "cover",
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
				PluginName:  "shell",
				Params: map[string]interface{}{
					"command": "rm -rf ./vendor",
				},
			},
		},
	},
}

// RunScaffoldManifest handles 'init' command
func RunScaffoldManifest(c *cli.Context) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current working directory, %s", err)
	}

	log.Default.Debugf("current working directory is '%s'", cwd)
	ensureGoPath(cwd)

	out, err := yaml.Marshal(boilerplate)
	if err != nil {
		return fmt.Errorf("cannot create YAML file: %s", err)
	}

	destFile := path.Join(cwd, manifest.FileName)
	err = ioutil.WriteFile(path.Join(cwd, manifest.FileName), out, 0655)
	if err != nil {
		return fmt.Errorf("failed to write '%s': %s", destFile, err)
	}

	log.Default.Successf("File '%s' successfully created", destFile)
	log.Default.Info("Use 'gilbert run build' to build the project")
	return nil
}

func ensureGoPath(cwd string) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Default.Warn("Warning: GOPATH environment variable is not defined")
		return
	}

	if !strings.Contains(cwd, goPath) {
		log.Default.Warnf("Current directory is outside GOPATH (%s)", goPath)
	}
}
