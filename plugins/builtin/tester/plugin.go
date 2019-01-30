package tester

import (
	"fmt"
	"github.com/x1unix/guru/env"
	"github.com/x1unix/guru/manifest"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/guru/logging"
	"github.com/x1unix/guru/plugins"
)

const allFiles = "*"

type Plugin struct {
	context *env.Context
	params  *Params
	console logging.Logger
	tempDir string
}

func (p *Plugin) runEntryTests(entry *TestEntry) (err error) {
	//if len(entry.Coverage.Ignore) > 0 {
	//	p.console.Warn("this plugin currently doesn't support ignore list, 'ignore' parameter will be ignored")
	//}
	//
	//cmd, _, err := entry.getTestingCommand(p.tempDir)
	//if err != nil {
	//	return err
	//}
	//
	//if entry.ShouldCheckCoverage() {
	//	out, err := cmd.Output()
	//} else {
	//	return cmd.Run()
	//}
	return nil
}

func (p *Plugin) Call() (err error) {
	p.tempDir, err = tempDir()
	if err != nil {
		return err
	}

	defer os.Remove(p.tempDir)
	//for _, testEntry := range p.params.Items {
	//	p.console.Log("- Running tests in '%s'", testEntry.Path)
	//	if err := p.runEntryTests(testEntry); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func NewPlugin(context *env.Context, rawParams manifest.RawParams, out logging.Logger) (plugins.Plugin, error) {
	p := &Plugin{
		console: out,
		context: context,
		params:  new(Params),
	}

	params := new(Params)
	if err := mapstructure.Decode(rawParams, params); err != nil {
		return nil, fmt.Errorf("invalid test configuration provided, %v", err)
	}

	for i, entry := range params.Items {
		fullTestPath, err := context.ExpandVariables(entry.Path)
		if err != nil {
			return nil, fmt.Errorf("cannot expose test path '%s': %v", entry.Path, err)
		}

		// Replace '*' as './...'
		fullTestPath = strings.Replace(fullTestPath, allFiles, "./...", -1)
		params.Items[i].Path = fullTestPath
	}

	p.params = params
	return p, nil
}
