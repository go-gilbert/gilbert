package runner

import (
	"context"
	"errors"
	"github.com/go-gilbert/gilbert/internal/runner/job"
	"github.com/go-gilbert/gilbert/internal/support/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTrackAsyncJobs(t *testing.T) {
	l := &test.Log{T: t}
	ctx := context.Background()
	tr := newAsyncJobTracker(ctx, l, 1)
	rtx := job.NewRunContext(ctx, nil, l)

	tr.decorateJobContext(rtx)
	go tr.trackAsyncJobs()

	go func() {
		time.Sleep(time.Millisecond * 300)
		rtx.Result(errors.New("foo"))
	}()

	assert.NoError(t, tr.wait())
}
