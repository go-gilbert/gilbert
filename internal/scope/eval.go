package scope

import (
	"fmt"
	"os/exec"

	"github.com/go-gilbert/gilbert/internal/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/support/shell"
)

var (
	_ expr.CommandProcessor = (*scopeExprAdapter)(nil)
	_ expr.ValueResolver    = (*scopeExprAdapter)(nil)
)

// scopeExprAdapter implements CommandProcessor and ValueResolver for expression parser to operate on a scope.
//
// This is a workaround type until expression parsing and evaluation won't be split.
type scopeExprAdapter struct {
	ctx *Scope
}

func (e scopeExprAdapter) prepareProcess(cmd string) (proc *exec.Cmd) {
	proc = shell.PrepareCommand(cmd)
	vars := shell.Environment(e.ctx.Variables)
	proc.Dir = e.ctx.environment.ProjectDirectory

	if !vars.Empty() {
		proc.Env = vars.ToArray(e.ctx.Environ()...)
	} else {
		proc.Env = e.ctx.Environ()
	}
	return proc
}

func (e scopeExprAdapter) EvalCommand(cmd string) (result []byte, err error) {
	proc := e.prepareProcess(cmd)

	data, err := proc.CombinedOutput()
	if err != nil {
		return result, fmt.Errorf("%w (%s)", shell.FormatExitError(err), data)
	}

	return data, nil
}

func (e scopeExprAdapter) ValueByName(varName string) (string, bool) {
	_, val, ok := e.ctx.Var(varName)
	return val, ok
}

func (e scopeExprAdapter) Values() any {
	// TODO: Will be replaced.
	return e.ctx.Variables
}

func (e scopeExprAdapter) evalContext() expr.EvalContext {
	return expr.EvalContext{
		CommandProcessor: e,
		Env:              e,
	}
}

func newScopeExprAdapter(ctx *Scope) scopeExprAdapter {
	return scopeExprAdapter{ctx: ctx}
}
