package engine

import (
	"context"
	"errors"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
)

type asyncJobGroup struct {
	runner         *Runner
	group          *errgroup.Group
	ctx            context.Context
	cancelFn       context.CancelFunc
	taskScope      *scope.Scope
	remainingCount atomic.Int32
}

func newAsyncJobGroup(r *Runner, cancelFn context.CancelFunc) *asyncJobGroup {
	return &asyncJobGroup{
		runner:   r,
		cancelFn: cancelFn,
	}
}

func (g *asyncJobGroup) isInitialized() bool {
	// initialized only if at-least one async job was scheduled
	return g.group != nil
}

func (g *asyncJobGroup) schedule(ctx context.Context, j manifest.Job, s *scope.Scope) {
	if g.group == nil {
		// lazy init on demand
		g.group, g.ctx = errgroup.WithContext(ctx)
	}

	g.remainingCount.Add(1)
	g.group.Go(func() error {
		defer g.remainingCount.Add(-1)
		// TODO: continue on timeout error
		if err := ctx.Err(); err != nil {
			return err
		}

		err := g.runner.runJob(g.ctx, j, s)
		if err != nil && !errors.Is(err, context.Canceled) {
			g.runner.logger.Named(j.Handler.String()).Error(err)
		}

		if j.ContinueOnError {
			return nil
		}

		g.cancelFn()
		return err
	})
}

func (g *asyncJobGroup) wait() error {
	return g.group.Wait()
}
