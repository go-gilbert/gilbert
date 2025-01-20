// +build windows

package shell

import (
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
func PrepareCommand(cmdName string) *exec.Cmd {
	cmd := exec.Command(shellPath, shellCmdPrefix, wrapCommand(cmdName))
	return cmd
}
