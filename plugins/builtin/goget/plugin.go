package goget

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/runner/job"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"
	"os/exec"
	"strings"
)

// Plugin implements gilbert plugin
type Plugin struct {
	scope   *scope.Scope
	params  params
	log     logging.Logger
	stopped bool
}

// Call implements plugins.plugin
func (p *Plugin) Call(ctx *job.RunContext, r plugins.JobRunner) error {
	if len(p.params.Packages) == 0 {
		return errors.New("no packages to install")
	}

	for _, pkg := range p.params.Packages {
		if p.stopped {
			return nil
		}

		if err := p.getPackage(pkg, ctx); err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) getPackage(pkgName string, ctx *job.RunContext) (err error) {
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

	go func() {
		select {
		case <-ctx.Context.Done():
			p.log.Debug("kill:", proc.Path)
			_ = proc.Process.Kill()
		}
	}()

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

func (p *Plugin) Cancel(ctx *job.RunContext) error {
	p.stopped = true
	return nil
}

// NewPlugin creates a new plugin instance
func NewPlugin(scope *scope.Scope, rawParams manifest.RawParams, log logging.Logger) (plugins.Plugin, error) {
	p := params{}
	if err := mapstructure.Decode(rawParams, &p); err != nil {
		return nil, manifest.NewPluginConfigError("build", err)
	}

	return &Plugin{
		scope:  scope,
		params: p,
		log:    log,
	}, nil
}
