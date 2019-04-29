package runner

import (
	"context"
	"sync"

	"github.com/go-gilbert/gilbert/runner/job"
)

// asyncJobTracker tracks state of async jobs
type asyncJobTracker struct {
	wg     *sync.WaitGroup
	errors chan error
	runner *TaskRunner
	ctx    context.Context
}

func newAsyncJobTracker(ctx context.Context, r *TaskRunner, poolSize int) *asyncJobTracker {
	return &asyncJobTracker{
		wg:     &sync.WaitGroup{},
		errors: make(chan error, poolSize),
		runner: r,
		ctx:    ctx,
	}
}

// trackAsyncJobs tracks errors from async jobs
func (t *asyncJobTracker) trackAsyncJobs() {
	select {
	case err, ok := <-t.errors:
		if ok && err != nil {
			t.runner.subLogger.Errorf("ERROR: async job returned error: %s", err)
		}
	case <-t.ctx.Done():
		return
	}
}

// decorateJobContext binds tracker to job context
func (t *asyncJobTracker) decorateJobContext(ctx *job.RunContext) {
	t.wg.Add(1)
	ctx.SetWaitGroup(t.wg)
	ctx.Error = t.errors
}

// wait waits until all async jobs complete
func (t *asyncJobTracker) wait() (err error) {
	// Wait for unfinished async tasks
	// and collect results from async jobs
	t.wg.Wait()
	close(t.errors)

	// TODO: report if any of async jobs failed
	return err
}
