package job

import (
	"context"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
	"sync"
	"time"
)

type RunContext struct {
	RootVars scope.Vars
	Logger   logging.Logger
	Context  context.Context
	Error    chan error
	wg       *sync.WaitGroup
	finished bool
}

func (r *RunContext) SetWaitGroup(wg *sync.WaitGroup) {
	r.wg = wg
}

func (r *RunContext) ChildContext() RunContext {
	return RunContext{
		RootVars: r.RootVars,
		Logger:   r.Logger.SubLogger(),
		Context:  r.Context,
	}
}

func (r *RunContext) WithTimeout(t time.Duration) RunContext {
	ctx, _ := context.WithTimeout(r.Context, t)
	return RunContext{
		RootVars: r.RootVars,
		Logger:   r.Logger,
		Context:  ctx,
	}
}

func (r *RunContext) Success() {
	r.Fail(nil)
}

func (r *RunContext) Fail(err error) {
	if r.finished {
		return
	}

	r.Error <- err
	r.finished = true
	if r.wg != nil {
		r.wg.Done()
	}
}

func NewRunContext(rootVars scope.Vars, log logging.Logger, ctx context.Context) RunContext {
	return RunContext{RootVars: rootVars, Logger: log, Context: ctx, Error: make(chan error)}
}
