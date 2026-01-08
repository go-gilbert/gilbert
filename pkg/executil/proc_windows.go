package executil

import "os/exec"

const (
	ShellPath      = "cmd.exe"
	ShellCmdOption = "/C"
)

func AddProcessGroup(_ *exec.Cmd) {
	// Windows doesn't support process groups
}

// KillProcessGroup kills process group created by parent process
func KillProcessGroup(cmd *exec.Cmd) {
	// Windows doesn't support process groups
}
