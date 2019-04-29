package shell

import (
	"fmt"
	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/tools/shell"
	"os"
	"os/exec"
	"strings"
)

// Plugin represents Gilbert's plugin
type Plugin struct {
	scope  sdk.ScopeAccessor
	params Params
	log    sdk.Logger
	cmd    *exec.Cmd
}

// Call calls a plugin
func (p *Plugin) Call(tx sdk.JobContextAccessor, r sdk.JobRunner) (err error) {
	p.cmd, err = p.params.createProcess(p.scope)
	if err != nil {
		return fmt.Errorf("failed to create process to execute command '%s': %s", p.params.Command, err)
	}

	p.log.Debugf("command: '%s'", p.params.Command)
	p.log.Debugf(`starting process "%s"...`, strings.Join(p.cmd.Args, " "))

	// Add std listeners when silent is off
	if !p.params.Silent {
		p.decorateProcessOutput(p.cmd)
	}

	if err = p.cmd.Start(); err != nil {
		return fmt.Errorf(`failed to execute command "%s": %s`, strings.Join(p.cmd.Args, " "), err)
	}

	if err = p.cmd.Wait(); err != nil {
		return formatExitError(err)
	}

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

// Cancel cancels shell command execution
func (p *Plugin) Cancel(ctx sdk.JobContextAccessor) error {
	if p.cmd == nil {
		return nil
	}

	// TODO: use exec.CommandContext to kill process
	p.log.Debug("received stop signal")
	if err := shell.KillProcessGroup(p.cmd); err != nil {
		p.log.Debugf("process killed with error: %s", err)
	}

	return nil
}
