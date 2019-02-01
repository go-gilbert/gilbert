package shell

import (
	"fmt"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
	"os"
	"os/exec"
	"strings"
)

type Plugin struct {
	context *scope.Context
	params  Params
	log     logging.Logger
}

func (p *Plugin) Call() error {
	proc, err := p.params.createProcess(p.context)
	if err != nil {
		return fmt.Errorf("failed to create process to execute command '%s': %s", p.params.Command, err)
	}

	p.log.Debug("command: '%s'", p.params.Command)
	p.log.Debug(`starting process "%s"...`, strings.Join(proc.Args, " "))

	// Add std listeners when silent is off
	if !p.params.Silent {
		p.decorateProcessOutput(proc)
	}

	if err = proc.Start(); err != nil {
		err = fmt.Errorf(`failed to execute command "%s": %s`, strings.Join(proc.Args, " "), err)
	}

	if err := proc.Wait(); err != nil {
		return formatExitError(err)
	}
	return nil
}

func (p *Plugin) decorateProcessOutput(proc *exec.Cmd) {
	proc.Stdin = os.Stdin

	if p.params.RawOutput {
		p.log.Debug("raw output enabled")
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr
		return
	}

	proc.Stdout = p.log
	proc.Stderr = p.log.ErrorWriter()
}
