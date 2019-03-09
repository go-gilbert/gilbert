package watch

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/runner/job"
	"github.com/x1unix/gilbert/scope"
)

type Plugin struct {
	params
	scope   *scope.Scope
	log     logging.Logger
	watcher *fsnotify.Watcher
	done    chan bool
}

func newPlugin(s *scope.Scope, p params, l logging.Logger) (*Plugin, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize watcher, %s", err)
	}

	return &Plugin{
		params:  p,
		scope:   s,
		log:     l,
		watcher: watcher,
		done:    make(chan bool, 1),
	}, nil
}

func (p *Plugin) Call(ctx *job.RunContext, r plugins.TaskRunner) (err error) {
	defer func() {
		if err = p.watcher.Close(); err != nil {
			p.log.Warn("failed to close watcher: %s", err)
		}
	}()

	go func() {
		for {
			select {
			case event, ok := <-p.watcher.Events:
				if !ok {
					return
				}
				p.log.Debug("event: %s", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					p.log.Info("modified file: %s", event.Name)
				}
			case err, ok := <-p.watcher.Errors:
				if !ok {
					return
				}
				p.log.Error("error: %s", err)
			}
		}
	}()

	err = p.watcher.Add(p.Path)
	if err != nil {
		return err
	}

	<-p.done
	return nil
}

func (p *Plugin) Cancel(ctx *job.RunContext) error {
	p.done <- true
	return nil
}
