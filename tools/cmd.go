package tools

import (
	"fmt"
	"os/exec"
	"syscall"
)

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
