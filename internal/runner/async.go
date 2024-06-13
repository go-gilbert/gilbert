package runner

import (
	"context"
	"sync"

	sdk "github.com/go-gilbert/gilbert-sdk"
)

// asyncJobTracker tracks state of async jobs
type asyncJobTracker struct {
	wg     *sync.WaitGroup
	errors chan error
	log    sdk.Logger
	ctx    context.Context
}

func newAsyncJobTracker(ctx context.Context, l sdk.Logger, poolSize int) *asyncJobTracker {
	return &asyncJobTracker{
		wg:     &sync.WaitGroup{},
		errors: make(chan error, poolSize),
		log:    l,
		ctx:    ctx,
	}
}

// trackAsyncJobs tracks errors from async jobs
func (t *asyncJobTracker) trackAsyncJobs() {
	select {
	case err, ok := <-t.errors:
		if ok && err != nil {
			t.log.Errorf("ERROR: async job returned error: %s", err)
		}
	case <-t.ctx.Done():
		return
	}
}

// decorateJobContext binds tracker to job context
func (t *asyncJobTracker) decorateJobContext(ctx sdk.JobContextAccessor) {
	t.wg.Add(1)
	ctx.SetWaitGroup(t.wg)
	ctx.SetErrorChannel(t.errors)
}

// wait waits until all async jobs complete
func (t *asyncJobTracker) wait() error {
	// Wait for unfinished async tasks
	// and collect results from async jobs
	t.wg.Wait()
	close(t.errors)

	// TODO: report if any of async jobs failed
	return nil
}
