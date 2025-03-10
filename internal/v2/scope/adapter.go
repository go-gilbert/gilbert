package scope

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/go-gilbert/gilbert/internal/support/shell"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/expr"
)

const commandEvalTimeout = 30 * time.Second

// CommandRunner implements CommandProcessor for expression parser to operate on a scope.
type CommandRunner struct {
	scope *Scope
}

func NewCommandRunner(scope *Scope) CommandRunner {
	return CommandRunner{
		scope: scope,
	}
}

func (runner CommandRunner) prepareProcess(ctx context.Context, cmd string) (proc *exec.Cmd) {
	proc = shell.PrepareContextCommand(ctx, cmd)
	vars := shell.Environment(runner.scope.Globals.Env)
	proc.Dir = runner.scope.Globals.Project.WorkDir()

	// TODO: should inputs & consts be exported into env?
	if !vars.Empty() {
		proc.Env = vars.ToArray()
	}

	return proc
}

func (runner CommandRunner) EvalCommand(ctx context.Context, cmdline string) (result []byte, err error) {
	cmdCtx, cancelFn := context.WithTimeout(ctx, commandEvalTimeout)
	defer cancelFn()

	proc := runner.prepareProcess(cmdCtx, cmdline)
	data, err := proc.CombinedOutput()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("command %q exceeded execution timeout %s", cmdline, commandEvalTimeout)
		}

		return result, fmt.Errorf("%w (%s)", shell.FormatExitError(err), data)
	}

	return data, nil
}

// NewEvalContext builds eval context for expanding dynamic expressions.
func NewEvalContext(s *Scope) expr.EvalContext {
	return expr.EvalContext{
		CommandProcessor: NewCommandRunner(s),
		Env:              s,
	}
}
