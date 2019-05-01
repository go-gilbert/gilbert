package pkgget

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/tools/shell"
)

// Action implements sdk.Action
type Action struct {
	scope   sdk.ScopeAccessor
	params  params
	log     sdk.Logger
	stopped bool
}

// Call implements plugins.plugin
func (a *Action) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) error {
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

func (a *Action) getPackage(pkgName string, ctx sdk.JobContextAccessor) (err error) {
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
		select {
		case <-ctx.Context().Done():
			a.log.Debug("kill:", proc.Path)
			_ = proc.Process.Kill()
		}
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
func (a *Action) Cancel(_ sdk.JobContextAccessor) error {
	a.stopped = true
	return nil
}

// NewAction creates a get-package action handler instance
func NewAction(scope sdk.ScopeAccessor, rawParams sdk.ActionParams) (sdk.ActionHandler, error) {
	p := params{}
	if err := rawParams.Unmarshal(&p); err != nil {
		return nil, err
	}

	return &Action{
		scope:  scope,
		params: p,
	}, nil
}
