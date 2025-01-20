package build

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-gilbert/gilbert/internal/runner"
	"github.com/go-gilbert/gilbert/internal/runner/job"
	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/go-gilbert/gilbert/internal/support/shell"
)

// Action represents Gilbert's plugin
type Action struct {
	scope  *scope.Scope
	cmd    *exec.Cmd
	params Params
}

// Call calls a plugin
func (a *Action) Call(ctx *job.RunContext, _ *runner.TaskRunner) (err error) {
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
		return shell.FormatExitError(err)
	}

	return nil
}

// Cancel cancels build process
func (a *Action) Cancel(ctx *job.RunContext) error {
	if a.cmd != nil {
		if err := shell.KillProcessGroup(a.cmd); err != nil {
			ctx.Log().Debug(err.Error())
		}
	}

	return nil
}
