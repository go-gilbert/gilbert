package shell

import (
	"os/exec"
	"syscall"
)

const (
	shellSh         = "/bin/sh"
	shCommandPrefix = "-c"
)

func defaultParams() Params {
	return Params{
		Shell:          shellSh,
		ShellExecParam: shCommandPrefix,
	}
}

func (p *Params) preparedCommand() string {
	return p.Command
}

func decorateCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
