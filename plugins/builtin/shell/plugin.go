package shell

import (
	"fmt"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/runner/job"
	"github.com/x1unix/gilbert/scope"
	"os"
	"os/exec"
	"strings"
)

// Plugin represents Gilbert's plugin
type Plugin struct {
	scope  *scope.Scope
	params Params
	log    logging.Logger
	proc   *exec.Cmd
}

// Call calls a plugin
func (p *Plugin) Call(tx *job.RunContext, r plugins.TaskRunner) (err error) {
	p.proc, err = p.params.createProcess(p.scope)
	if err != nil {
		return fmt.Errorf("failed to create process to execute command '%s': %s", p.params.Command, err)
	}

	p.log.Debug("command: '%s'", p.params.Command)
	p.log.Debug(`starting process "%s"...`, strings.Join(p.proc.Args, " "))

	// Add std listeners when silent is off
	if !p.params.Silent {
		p.decorateProcessOutput()
	}

	if err = p.proc.Start(); err != nil {
		return fmt.Errorf(`failed to execute command "%s": %s`, strings.Join(p.proc.Args, " "), err)
	}

	if err := p.proc.Wait(); err != nil {
		return formatExitError(err)
	}

	p.log.Debug("done")
	return nil
}

func (p *Plugin) decorateProcessOutput() {
	if p.params.RawOutput {
		p.log.Debug("raw output enabled")
		p.proc.Stdout = os.Stdout
		p.proc.Stderr = os.Stderr
		p.proc.Stdin = os.Stdin
		return
	}

	p.proc.Stdout = p.log
	p.proc.Stderr = p.log.ErrorWriter()
}

func (p *Plugin) Cancel(ctx *job.RunContext) error {
	if p.proc != nil {
		p.log.Debug("received stop signal")
		if err := p.proc.Process.Kill(); err != nil {
			p.log.Warn(err.Error())
		}
	}

	return nil
}
