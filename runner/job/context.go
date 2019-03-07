package job

import (
	"context"
	"sync"
	"time"

	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
)

type RunContext struct {
	RootVars scope.Vars
	Logger   logging.Logger
	Context  context.Context
	Error    chan error
	wg       *sync.WaitGroup
	once     sync.Once
}

func (r *RunContext) SetWaitGroup(wg *sync.WaitGroup) {
	r.wg = wg
}

func (r *RunContext) ChildContext() RunContext {
	return RunContext{
		RootVars: r.RootVars,
		Logger:   r.Logger.SubLogger(),
		Context:  r.Context,
		Error:    r.Error,
		wg:       r.wg,
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
	r.once.Do(func() {
		defer func() {
			if rec := recover(); rec != nil {
				r.Logger.Error("Bug: failed to return job result, %v", rec)
				r.wg.Done()
			}
		}()

		r.Error <- err
		if r.wg != nil {
			r.Logger.Debug("ctx: done reported")
			r.wg.Done()
		}
	})
}

func NewRunContext(rootVars scope.Vars, log logging.Logger, ctx context.Context) RunContext {
	return RunContext{RootVars: rootVars, Logger: log, Context: ctx, Error: make(chan error, 1)}
}
