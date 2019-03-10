package shell

import (
	"os/exec"
	"syscall"
)

const (
	shellPath      = "/bin/sh"
	shellCmdPrefix = "-c"
)

func wrapCommand(cmd string) string {
	return cmd
}

// KillProcessGroup kills process group created by parent process
func KillProcessGroup(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

// PrepareCommand prepares a command to execute
func PrepareCommand(cmdName string) *exec.Cmd {
	cmd := exec.Command(shellPath, shellCmdPrefix, wrapCommand(cmdName))

	// Assign process group (for unix only)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}
