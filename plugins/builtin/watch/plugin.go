package watch

import (
	"fmt"
	"sync"
	"time"

	"github.com/rjeczalik/notify"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/runner/job"
	"github.com/x1unix/gilbert/scope"
)

type Plugin struct {
	params
	scope  *scope.Scope
	log    logging.Logger
	done   chan bool
	events chan notify.EventInfo
	dead   *sync.Mutex
}

func newPlugin(s *scope.Scope, p params, l logging.Logger) (*Plugin, error) {
	return &Plugin{
		params: p,
		scope:  s,
		log:    l,
		done:   make(chan bool),
	}, nil
}

func (p *Plugin) Call(ctx *job.RunContext, r plugins.JobRunner) error {
	p.events = make(chan notify.EventInfo, 1)
	if err := notify.Watch(p.Path, p.events, notify.All); err != nil {
		return fmt.Errorf("failed to initialize watcher for '%s': %s", p.Path, err)
	}

	p.dead = &sync.Mutex{}
	childCtx := ctx.ChildContext()
	defer func() {
		notify.Stop(p.events)
		childCtx.Cancel()
		p.log.Debug("watcher removed")
	}()

	// Start file watcher
	go func() {
		interval := p.DebounceTime.ToDuration()
		timer := time.NewTimer(interval) // Debounce timer

		for {
			select {
			case event, ok := <-p.events:
				if !ok {
					return
				}
				// Reset timer when new FS event came
				p.log.Debug("event: %v %s", event.Event(), event.Path())
				timer.Reset(interval)
			case <-timer.C:
				// Re-start job when timer ends.
				p.log.Debug("timer ended")

				if childCtx.IsAlive() {
					childCtx.Cancel()
				}
				childCtx = ctx.ChildContext()
				go p.invokeJob(childCtx, r)
			}
		}
	}()

	p.log.Info("watcher is watching for changes in '%s'", p.Path)
	<-p.done
	return nil
}

func (p *Plugin) invokeJob(ctx job.RunContext, r plugins.JobRunner) {
	p.log.Debug("wait until previous process stops")
	p.dead.Lock()
	ctx.Error = make(chan error, 1)
	p.log.Info("- Starting '%s'", p.Job.FormatDescription())
	r.RunJob(p.Job, ctx)
	select {
	case err := <-ctx.Error:
		p.log.Debug("tracked job finished")
		p.dead.Unlock()
		if err != nil {
			p.log.Error("Job failed: %s", err)
		}
	}
}

func (p *Plugin) Cancel(ctx *job.RunContext) error {
	p.done <- true
	notify.Stop(p.events)
	p.log.Debug("watcher removed")
	return nil
}
