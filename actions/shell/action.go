package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/support/shell"
)

// Action represents sdk.ActionHandler
type Action struct {
	scope  sdk.ScopeAccessor
	params Params
	cmd    *exec.Cmd
}

// Call calls a plugin
func (a *Action) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) (err error) {
	a.cmd, err = a.params.createProcess(a.scope)
	if err != nil {
		return fmt.Errorf("failed to create process to execute command '%s': %s", a.params.Command, err)
	}

	ctx.Log().Debugf("command: '%s'", a.params.Command)
	ctx.Log().Debugf(`starting process "%s"...`, strings.Join(a.cmd.Args, " "))

	// Add std listeners when silent is off
	if !a.params.Silent {
		a.decorateProcessOutput(ctx, a.cmd)
	}

	if err = a.cmd.Start(); err != nil {
		return fmt.Errorf(`failed to execute command "%s": %s`, strings.Join(a.cmd.Args, " "), err)
	}

	if err = a.cmd.Wait(); err != nil {
		return formatExitError(err)
	}

	return nil
}

func (a *Action) decorateProcessOutput(ctx sdk.JobContextAccessor, cmd *exec.Cmd) {
	if a.params.RawOutput {
		ctx.Log().Debug("raw output enabled")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return
	}

	cmd.Stdout = ctx.Log()
	cmd.Stderr = ctx.Log().ErrorWriter()
}

// Cancel cancels shell command execution
func (a *Action) Cancel(ctx sdk.JobContextAccessor) error {
	if a.cmd == nil {
		return nil
	}

	// TODO: use exec.CommandContext to kill process
	ctx.Log().Debug("received stop signal")
	if err := shell.KillProcessGroup(a.cmd); err != nil {
		ctx.Log().Debugf("process killed with error: %s", err)
	}

	return nil
}
