package build

import (
	"fmt"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/runner/job"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"
	"os/exec"
	"strings"
)

// Plugin represents Gilbert's plugin
type Plugin struct {
	scope  *scope.Scope
	cmd    *exec.Cmd
	params Params
	log    logging.Logger
}

// Call calls a plugin
func (p *Plugin) Call(ctx *job.RunContext, r plugins.TaskRunner) (err error) {
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

	ctx.Logger.Debug("started")
	if err = p.cmd.Wait(); err != nil {
		return shell.FormatExitError(err)
	}

	ctx.Logger.Debug("waited")
	return nil
}

func (p *Plugin) Cancel(ctx *job.RunContext) error {
	if p.cmd != nil {
		if err := p.cmd.Process.Kill(); err != nil {
			p.log.Warn(err.Error())
		}
	}

	return nil
}
