//go:build windows
// +build windows

package shell

import (
	"context"
	"os/exec"
)

const (
	shellPath            = "cmd.exe"
	shellCmdPrefix       = "/C"
	winCodePageFixPrefix = "chcp 65001 > nul" // Force use UTF-8 to provide correct output to stdout
)

func wrapCommand(cmd string) string {
	return winCodePageFixPrefix + " && " + cmd
}

// KillProcessGroup kills process group created by parent process
func KillProcessGroup(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}

// PrepareCommand prepares a command to execute
//
// Deprecated: use PrepareContextCommand instead.
func PrepareCommand(cmdName string) *exec.Cmd {
	cmd := exec.Command(shellPath, shellCmdPrefix, wrapCommand(cmdName))
	return cmd
}

// PrepareContextCommand prepares a command to execute
func PrepareContextCommand(ctx context.Context, cmdName string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, shellPath, shellCmdPrefix, wrapCommand(cmdName))
	return cmd
}
