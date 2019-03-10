package job

import (
	"context"
	"sync"
	"time"

	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
)

// RunContext used to store job state and communicate between task runner and job
type RunContext struct { // nolint: maligned
	child    bool
	finished bool

	// Logger is sub-logger instance for the job
	Logger logging.Logger

	// Context is context.Context instance for current job.
	Context context.Context

	// Error is job result channel
	Error chan error

	wg       *sync.WaitGroup
	cancelFn context.CancelFunc
	once     sync.Once

	// RootVars used to hold variables of root context
	RootVars scope.Vars
}

// SetWaitGroup sets wait group instance for current job
//
// This value will be used later to call wg.Done() when job was finished.
func (r *RunContext) SetWaitGroup(wg *sync.WaitGroup) {
	r.wg = wg
}

// IsChild checks if context is child context
func (r *RunContext) IsChild() bool {
	return r.child
}

// ForkContext creates a context copy, but creates a separate sub-logger
func (r *RunContext) ForkContext() RunContext {
	return RunContext{
		RootVars: r.RootVars,
		Logger:   r.Logger.SubLogger(),
		Context:  r.Context,
		Error:    r.Error,
		cancelFn: r.cancelFn,
		child:    true,
		wg:       r.wg,
	}
}

// ChildContext creates a new child context with separate Error channel and context
func (r *RunContext) ChildContext() *RunContext {
	ctx, cancelFn := context.WithCancel(r.Context)

	return &RunContext{
		RootVars: r.RootVars,
		Logger:   r.Logger.SubLogger(),
		Context:  ctx,
		Error:    make(chan error, 1),
		cancelFn: cancelFn,
		child:    true,
	}
}

// Timeout adds timeout to the context
func (r *RunContext) Timeout(timeout time.Duration) {
	// Used as workaround, since context.WithTimeout() reports Done()
	// but hangs goroutine.
	go func() {
		select {
		case <-time.After(timeout):
			if !r.finished {
				r.Logger.Warn("Job deadline exceeded")
				r.Cancel()
			}
		case <-r.Context.Done():
			return
		}
	}()
}

// Cancel cancels the context and stops all jobs used by this context
func (r *RunContext) Cancel() {
	r.finished = true
	if r.cancelFn == nil {
		r.Logger.Error("Bug: context cancel function is nil")
		return
	}

	r.cancelFn()
}

// Success reports successful result.
//
// Alias to Result(nil)
func (r *RunContext) Success() {
	r.Result(nil)
}

// IsAlive checks if context was not finished
func (r *RunContext) IsAlive() bool {
	return !r.finished
}

// Result reports job result and finished the context
func (r *RunContext) Result(err error) {
	r.once.Do(func() {
		defer func() {
			if rec := recover(); rec != nil {
				r.Logger.Error("Bug: failed to return job result, %v", rec)
				if r.wg != nil {
					r.wg.Done()
				}
			}
		}()

		r.finished = true
		r.Error <- err
		r.Logger.Debug("result received")
		if r.wg != nil {
			r.wg.Done()
		}
	})
}

// NewRunContext creates a new job context instance
func NewRunContext(parentCtx context.Context, rootVars scope.Vars, log logging.Logger) *RunContext {
	ctx, cancelFn := context.WithCancel(parentCtx)
	return &RunContext{RootVars: rootVars, Logger: log, Context: ctx, Error: make(chan error, 1), cancelFn: cancelFn}
}
