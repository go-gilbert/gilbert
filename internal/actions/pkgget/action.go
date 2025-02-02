package pkgget

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"
	"github.com/go-gilbert/gilbert/internal/runner"
	"github.com/go-gilbert/gilbert/internal/runner/job"
	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/go-gilbert/gilbert/internal/support/shell"
)

type Action struct {
	scope   *scope.Scope
	params  params
	log     log.Logger
	stopped bool
}

// Call implements plugins.plugin
func (a *Action) Call(ctx *job.RunContext, r *runner.TaskRunner) error {
	if len(a.params.Packages) == 0 {
		return errors.New("no packages to install")
	}

	for _, pkg := range a.params.Packages {
		if a.stopped {
			return nil
		}

		if err := a.getPackage(pkg, ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *Action) getPackage(pkgName string, ctx *job.RunContext) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get package '%s', %s", pkgName, err)
		}
	}()

	cmd := []string{"get"}
	if a.params.DownloadOnly {
		cmd = append(cmd, "-d")
	}

	if a.params.Update {
		cmd = append(cmd, "-u")
	}

	if a.params.Verbose {
		cmd = append(cmd, "-v")
	}

	cmd = append(cmd, pkgName)
	proc := exec.Command("go", cmd...)
	if a.params.Verbose {
		proc.Stdout = a.log
	}

	go func() {
		<-ctx.Context().Done()
		a.log.Debugf("get-package: kill '%s'", proc.Path)
		_ = proc.Process.Kill()
	}()

	a.log.Infof("Downloading package '%s'", pkgName)
	a.log.Debug(strings.Join(proc.Args, " "))
	if err := proc.Start(); err != nil {
		return err
	}

	if err = proc.Wait(); err != nil {
		return shell.FormatExitError(err)
	}

	return err
}

// Cancel cancels plugin execution
func (a *Action) Cancel(_ *job.RunContext) error {
	a.stopped = true
	return nil
}

// NewAction creates a get-package action handler instance
func NewAction(scope *scope.Scope, rawParams manifest.ActionParams) (runner.ActionHandler, error) {
	p := params{}
	if err := rawParams.Unmarshal(&p); err != nil {
		return nil, err
	}

	return &Action{
		scope:  scope,
		params: p,
	}, nil
}
