package scope

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/go-gilbert/gilbert/internal/support/shell"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/expr"
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

func (e scopeExprAdapter) EvalCommand(_ context.Context, cmd string) (result []byte, err error) {
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

func (e scopeExprAdapter) Values() map[string]any {
	// FIXME: keep this to get v1 building. remove when v1 is decommissioned.
	dst := make(map[string]any, len(e.ctx.Variables))
	for k, v := range e.ctx.Variables {
		dst[k] = v
	}

	return dst
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
