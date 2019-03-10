package shell

import (
	"fmt"
	"os/exec"
	"runtime"
	"syscall"
)

// OsWindows is windows os name
const OsWindows = "windows"

// FormatExitError extracts process wait error and formats it
func FormatExitError(err error) error {
	if exiterr, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0

		// This works on both Unix and Windows. Although package
		// syscall is generally platform dependent, WaitStatus is
		// defined for both Unix and Windows and in both cases has
		// an ExitStatus() method with the same signature.
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return fmt.Errorf("process finished with non-zero status code: %d", status.ExitStatus())
		}
	}

	return fmt.Errorf("process finished with error - %s", err)
}

// PrepareCommand prepares a command to execute
func PrepareCommand(cmdName string) *exec.Cmd {
	cmd := exec.Command(shellPath, shellCmdPrefix, wrapCommand(cmdName))

	// Assign process group (for unix only)
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	return cmd
}

// KillProcessGroup kills process group created by parent process
func KillProcessGroup(cmd *exec.Cmd) error {
	if runtime.GOOS == OsWindows {
		return cmd.Process.Kill()
	}

	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
