package goget

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"
	"os/exec"
	"strings"
)

// Plugin implements gilbert plugin
type Plugin struct {
	context *scope.Context
	params  params
	log     logging.Logger
}

// Call implements plugins.plugin
func (p *Plugin) Call() error {
	if len(p.params.Packages) == 0 {
		return errors.New("no packages to install")
	}

	for _, pkg := range p.params.Packages {
		if err := p.getPackage(pkg); err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) getPackage(pkgName string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get package '%s', %s", pkgName, err)
		}
	}()

	cmd := []string{"get"}
	if p.params.DownloadOnly {
		cmd = append(cmd, "-d")
	}

	if p.params.Update {
		cmd = append(cmd, "-u")
	}

	if p.params.Verbose {
		cmd = append(cmd, "-v")
	}

	cmd = append(cmd, pkgName)
	proc := exec.Command("go", cmd...)
	if p.params.Verbose {
		proc.Stdout = p.log
	}

	p.log.Info("Downloading package '%s'", pkgName)
	p.log.Debug(strings.Join(proc.Args, " "))
	if err := proc.Start(); err != nil {
		return err
	}

	if err = proc.Wait(); err != nil {
		return shell.FormatExitError(err)
	}

	return err
}

// NewPlugin creates a new plugin instance
func NewPlugin(context *scope.Context, rawParams manifest.RawParams, log logging.Logger) (plugins.Plugin, error) {
	p := params{}
	if err := mapstructure.Decode(rawParams, &p); err != nil {
		return nil, manifest.NewPluginConfigError("build", err)
	}

	return &Plugin{
		context: context,
		params:  p,
		log:     log,
	}, nil
}
