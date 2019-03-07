package runner

import (
	"context"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
	"sync"
	"time"
)

type JobContext struct {
	TaskName string
	Step     int
	RootVars scope.Vars
	Logger   logging.Logger
	Context  context.Context
	Error    chan error
	wg       *sync.WaitGroup
	finished bool
}

func (c *JobContext) SetMetadata(taskName string, step int) {
	c.TaskName = taskName
	c.Step = step
}

func (c *JobContext) ChildContext() JobContext {
	return JobContext{
		RootVars: c.RootVars,
		Logger:   c.Logger.SubLogger(),
		Context:  c.Context,
	}
}

func (c *JobContext) WithTimeout(t time.Duration) JobContext {
	ctx, _ := context.WithTimeout(c.Context, t)
	return JobContext{
		RootVars: c.RootVars,
		Logger:   c.Logger,
		Context:  ctx,
	}
}

func (c *JobContext) Success() {
	c.Fail(nil)
}

func (c *JobContext) Fail(err error) {
	if c.finished {
		return
	}

	c.Error <- err
	c.finished = true
	if c.wg != nil {
		c.wg.Done()
	}
}

func NewJobContext(rootVars scope.Vars, log logging.Logger, ctx context.Context) JobContext {
	return JobContext{RootVars: rootVars, Logger: log, Context: ctx, Error: make(chan error)}
}
