package watch

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-gilbert/gilbert-sdk"

	"github.com/go-gilbert/gilbert/internal/runner/job"
	"github.com/rjeczalik/notify"
)

// Action implements plugins.Action interface
type Action struct {
	params
	scope  sdk.ScopeAccessor
	done   chan bool
	events chan notify.EventInfo
	dead   *sync.Mutex
}

// Call starts watch plugin
func (a *Action) Call(ctx sdk.JobContextAccessor, r sdk.JobRunner) error {
	a.events = make(chan notify.EventInfo, 1)
	if err := notify.Watch(a.Path, a.events, notify.All); err != nil {
		return fmt.Errorf("failed to initialize watcher for '%s': %s", a.Path, err)
	}

	a.dead = &sync.Mutex{}
	childCtx := ctx.ChildContext()
	defer func() {
		notify.Stop(a.events)
		childCtx.Cancel()
		ctx.Log().Debug("watch: watcher removed")
	}()

	// Start file watcher
	go func() {
		interval := a.DebounceTime.ToDuration()
		t := time.NewTimer(interval) // Debounce timer

		for {
			select {
			case event, ok := <-a.events:
				if !ok {
					return
				}
				fPath := event.Path()
				ignored, err := a.pathIgnored(fPath)
				if err != nil {
					ctx.Log().Errorf("path ignore check failed: %s", err)
					continue
				}

				if !ignored {
					ctx.Log().Debugf("watch: received event - %v %s", event.Event(), fPath)
					t.Reset(interval)
				}
			case <-t.C:
				// Re-start job when timer ends.
				ctx.Log().Debug("watch: timer ended")

				if childCtx.IsAlive() {
					childCtx.Cancel()
				}
				childCtx = ctx.ChildContext()
				go a.invokeJob(childCtx, r)
			}
		}
	}()

	ctx.Log().Infof("watcher is watching for changes in '%s'", a.Path)
	<-a.done
	return nil
}

func (a *Action) invokeJob(ctx sdk.JobContextAccessor, r sdk.JobRunner) {
	ctx.Log().Debug("watch: wait until previous process stops")
	a.dead.Lock()
	// override errors channel
	jctx := ctx.(*job.RunContext)
	jctx.Error = make(chan error, 1)

	description := a.Job.FormatDescription()
	ctx.Log().Infof("- Starting '%s'", description)
	r.RunJob(*a.Job, ctx)

	err := <-ctx.Errors()
	a.dead.Unlock()
	if err != nil {
		ctx.Log().Errorf("- '%s' failed: %s", description, err)
		return
	}

	ctx.Log().Successf("- '%s' finished", description)
}

// Cancel stops watch plugin
func (a *Action) Cancel(ctx sdk.JobContextAccessor) error {
	a.done <- true
	notify.Stop(a.events)
	ctx.Log().Debug("watch: watcher removed")
	return nil
}
