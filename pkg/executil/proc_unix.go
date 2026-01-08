//go:build !windows

package executil

import (
	"os/exec"
	"syscall"
)

const (
	ShellPath      = "/usr/bin/sh"
	ShellCmdOption = "-c"
)

// AddProcessGroup adds process group attribute to [exec.Cmd].
func AddProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// KillProcessGroup kills process group created by parent process
func KillProcessGroup(cmd *exec.Cmd) {
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
