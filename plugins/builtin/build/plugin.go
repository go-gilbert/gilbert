package build

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/runner/job"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"
)

// Plugin represents Gilbert's plugin
type Plugin struct {
	scope  *scope.Scope
	cmd    *exec.Cmd
	params Params
	log    log.Logger
}

// Call calls a plugin
func (p *Plugin) Call(ctx *job.RunContext, r plugins.JobRunner) (err error) {
	p.cmd, err = p.params.newCompilerProcess(p.scope)
	if err != nil {
		return err
	}

	p.log.Debug("Target: %s %s", p.params.Target.Os, p.params.Target.Arch)
	p.log.Debug("Command: '%s'", strings.Join(p.cmd.Args, " "))
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
func (p *Plugin) Cancel(ctx *job.RunContext) error {
	if p.cmd != nil {
		if err := shell.KillProcessGroup(p.cmd); err != nil {
			p.log.Debug(err.Error())
		}
	}

	return nil
}
