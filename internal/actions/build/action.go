package build

import (
	"fmt"
	shell2 "github.com/go-gilbert/gilbert/pkg/support/shell"
	"os/exec"
	"strings"

	"github.com/go-gilbert/gilbert-sdk"
)

// Action represents Gilbert's plugin
type Action struct {
	scope  sdk.ScopeAccessor
	cmd    *exec.Cmd
	params Params
}

// Call calls a plugin
func (a *Action) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) (err error) {
	a.cmd, err = a.params.newCompilerProcess(a.scope)
	if err != nil {
		return err
	}

	ctx.Log().Debugf("build: target: %s %s", a.params.Target.Os, a.params.Target.Arch)
	ctx.Log().Debugf("build: exec: '%s'", strings.Join(a.cmd.Args, " "))
	a.cmd.Stdout = ctx.Log()
	a.cmd.Stderr = ctx.Log().ErrorWriter()

	if err := a.cmd.Start(); err != nil {
		return fmt.Errorf(`failed build project, %s`, err)
	}

	if err = a.cmd.Wait(); err != nil {
		return shell2.FormatExitError(err)
	}

	return nil
}

// Cancel cancels build process
func (a *Action) Cancel(ctx sdk.JobContextAccessor) error {
	if a.cmd != nil {
		if err := shell2.KillProcessGroup(a.cmd); err != nil {
			ctx.Log().Debug(err.Error())
		}
	}

	return nil
}
