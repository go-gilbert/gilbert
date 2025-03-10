//go:build !windows && !js && !nacl

package shell

import (
	"context"
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

// PrepareCommand prepares a command to execute.
//
// Deprecated: use PrepareContextCommand instead.
func PrepareCommand(cmdName string) *exec.Cmd {
	cmd := exec.Command(shellPath, shellCmdPrefix, wrapCommand(cmdName))
	prepareExecCmd(cmd)

	return cmd
}

// PrepareContextCommand prepares a command to execute
func PrepareContextCommand(ctx context.Context, cmdName string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, shellPath, shellCmdPrefix, wrapCommand(cmdName))
	prepareExecCmd(cmd)

	return cmd
}

func prepareExecCmd(cmd *exec.Cmd) {
	// Assign process group (for unix only)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
