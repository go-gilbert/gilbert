package shell

import (
	"fmt"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/runner/job"
	"github.com/x1unix/gilbert/scope"
	"github.com/x1unix/gilbert/tools/shell"
	"os"
	"os/exec"
	"strings"
)

// Plugin represents Gilbert's plugin
type Plugin struct {
	scope  *scope.Scope
	params Params
	log    logging.Logger
	done   chan bool
}

// Call calls a plugin
func (p *Plugin) Call(tx *job.RunContext, r plugins.JobRunner) (err error) {
	defer close(p.done)
	cmd, err := p.params.createProcess(p.scope)
	if err != nil {
		return fmt.Errorf("failed to create process to execute command '%s': %s", p.params.Command, err)
	}

	p.log.Debug("command: '%s'", p.params.Command)
	p.log.Debug(`starting process "%s"...`, strings.Join(cmd.Args, " "))

	// Add std listeners when silent is off
	if !p.params.Silent {
		p.decorateProcessOutput(cmd)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf(`failed to execute command "%s": %s`, strings.Join(cmd.Args, " "), err)
	}

	go func() {
		select {
		case <-p.done:
			p.log.Debug("received stop signal")
			if err := shell.KillProcessGroup(cmd); err != nil {
				p.log.Warn("process killed with error: %s", err)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return formatExitError(err)
	}

	p.log.Debug("done")
	return nil
}

func (p *Plugin) decorateProcessOutput(cmd *exec.Cmd) {
	if p.params.RawOutput {
		p.log.Debug("raw output enabled")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return
	}

	cmd.Stdout = p.log
	cmd.Stderr = p.log.ErrorWriter()
}

func (p *Plugin) Cancel(ctx *job.RunContext) error {
	p.done <- true
	return nil
}
