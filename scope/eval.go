package scope

import (
	"fmt"
	"github.com/x1unix/gilbert/tools/shell"
	"os/exec"
)

// CommandEvaluator represents command runner and wraps shell calls
type CommandEvaluator interface {
	// Call executes a shell command
	Call(string) ([]byte, error)
}

type shellEvaluator struct {
	ctx *Scope
}

func (e *shellEvaluator) prepareProcess(cmd string) (proc *exec.Cmd) {
	proc = shell.PrepareCommand(cmd)
	vars := shell.Environment(e.ctx.Variables)
	proc.Dir = e.ctx.Environment.ProjectDirectory

	if !vars.Empty() {
		proc.Env = vars.ToArray(e.ctx.Environ()...)
	} else {
		proc.Env = e.ctx.Environ()
	}
	return proc
}

// Call executes a shell command
func (e *shellEvaluator) Call(cmd string) (result []byte, err error) {
	proc := e.prepareProcess(cmd)

	data, err := proc.CombinedOutput()
	if err != nil {
		return result, fmt.Errorf("%s (%s)", shell.FormatExitError(err), data)
	}

	return data, nil
}

func newShellCommandEvaluator(ctx *Scope) CommandEvaluator {
	return &shellEvaluator{ctx: ctx}
}
