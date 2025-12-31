package executil

import "os/exec"

const (
	ShellPath      = "cmd.exe"
	ShellCmdOption = "/C"
)

func AddProcessGroup(_ *exec.Cmd) {
	// Windows doesn't support process groups
}
