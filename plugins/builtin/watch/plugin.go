package watch

import (
	"fmt"
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
}

func newPlugin(s *scope.Scope, p params, l logging.Logger) (*Plugin, error) {
	return &Plugin{
		params: p,
		scope:  s,
		log:    l,
		done:   make(chan bool),
	}, nil
}

func (p *Plugin) Call(ctx *job.RunContext, r plugins.TaskRunner) error {
	p.events = make(chan notify.EventInfo, 1)
	if err := notify.Watch(p.Path, p.events, notify.All); err != nil {
		return fmt.Errorf("failed to initialize watcher for '%s': %s", p.Path, err)
	}

	defer func() {
		notify.Stop(p.events)
		p.log.Debug("watcher removed")
	}()

	go func() {
		for {
			select {
			case event, ok := <-p.events:
				if !ok {
					return
				}
				p.log.Info("event: %v %s", event.Event(), event.Path())
			}
		}
	}()

	p.log.Info("watcher is watching for changes in '%s'", p.Path)
	<-p.done
	return nil
}

func (p *Plugin) Cancel(ctx *job.RunContext) error {
	p.done <- true
	notify.Stop(p.events)
	p.log.Debug("watcher removed")
	return nil
}
