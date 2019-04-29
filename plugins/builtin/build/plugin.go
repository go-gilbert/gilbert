package build

import (
	"fmt"
	"github.com/go-gilbert/gilbert-sdk"
	"os/exec"
	"strings"

	"github.com/x1unix/gilbert/tools/shell"
)

// Plugin represents Gilbert's plugin
type Plugin struct {
	scope  sdk.ScopeAccessor
	cmd    *exec.Cmd
	params Params
	log    sdk.Logger
}

// Call calls a plugin
func (p *Plugin) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) (err error) {
	p.cmd, err = p.params.newCompilerProcess(p.scope)
	if err != nil {
		return err
	}

	p.log.Debugf("Target: %s %s", p.params.Target.Os, p.params.Target.Arch)
	p.log.Debugf("Command: '%s'", strings.Join(p.cmd.Args, " "))
	p.cmd.Stdout = p.log
	p.cmd.Stderr = p.log.ErrorWriter()

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf(`failed build project, %s`, err)
	}

	if err = p.cmd.Wait(); err != nil {
		return shell.FormatExitError(err)
	}

	return nil
}

// Cancel cancels build process
func (p *Plugin) Cancel(ctx sdk.JobContextAccessor) error {
	if p.cmd != nil {
		if err := shell.KillProcessGroup(p.cmd); err != nil {
			p.log.Debug(err.Error())
		}
	}

	return nil
}
