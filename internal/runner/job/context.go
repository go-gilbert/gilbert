package job

import (
	"context"
	"sync"
	"time"

	"github.com/go-gilbert/gilbert/internal/log"
	"github.com/go-gilbert/gilbert/internal/manifest"
)

// RunContext used to store job state and communicate between task runner and job
type RunContext struct { // nolint: maligned
	child    bool
	finished bool

	// logger is sub-logger instance for the job
	logger log.Logger

	// context is context.context instance for current job.
	context context.Context

	// Error is job result channel
	Error chan error

	wg       *sync.WaitGroup
	cancelFn context.CancelFunc
	once     sync.Once

	// RootVars used to hold variables of root context
	RootVars manifest.Vars
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

// Errors returns job errors channel
func (r *RunContext) Errors() chan error {
	return r.Error
}

// SetErrorChannel sets custom error report channel
func (r *RunContext) SetErrorChannel(ch chan error) {
	r.Error = ch
}

// Log provides logger for current job context
func (r *RunContext) Log() log.Logger {
	return r.logger
}

// ForkContext creates a context copy, but creates a separate sub-logger
func (r *RunContext) ForkContext() *RunContext {
	return &RunContext{
		RootVars: r.RootVars,
		logger:   r.logger.SubLogger(),
		context:  r.context,
		Error:    r.Error,
		cancelFn: r.cancelFn,
		child:    true,
		wg:       r.wg,
	}
}

// Vars returns a set of variables attached to this context
func (r *RunContext) Vars() manifest.Vars {
	return r.RootVars
}

// SetVars sets context variables
func (r *RunContext) SetVars(newVars manifest.Vars) {
	r.RootVars = newVars
}

// ChildContext creates a new child context with separate Error channel and context
func (r *RunContext) ChildContext() *RunContext {
	ctx, cancelFn := context.WithCancel(r.context)

	return &RunContext{
		RootVars: r.RootVars,
		logger:   r.logger.SubLogger(),
		context:  ctx,
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
				r.logger.Warn("Job deadline exceeded")
				r.Cancel()
			}
		case <-r.context.Done():
			return
		}
	}()
}

// Cancel cancels the context and stops all jobs used by this context
func (r *RunContext) Cancel() {
	r.finished = true
	if r.cancelFn == nil {
		r.logger.Warn("Bug: context cancel function is nil")
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

// Context returns Go context instance assigned to the current job context
func (r *RunContext) Context() context.Context {
	return r.context
}

// Result reports job result and finished the context
func (r *RunContext) Result(err error) {
	r.once.Do(func() {
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Warnf("Bug: failed to return job result, %v", rec)
				if r.wg != nil {
					r.wg.Done()
				}
			}
		}()

		r.finished = true
		r.Error <- err
		r.logger.Debug("job: result received")
		if r.wg != nil {
			r.wg.Done()
		}
	})
}

// NewRunContext creates a new job context instance
func NewRunContext(parentCtx context.Context, rootVars manifest.Vars, l log.Logger) *RunContext {
	ctx, cancelFn := context.WithCancel(parentCtx)
	return &RunContext{RootVars: rootVars, logger: l, context: ctx, Error: make(chan error, 1), cancelFn: cancelFn}
}
