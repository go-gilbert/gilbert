package runner

import (
	"context"
	"github.com/x1unix/gilbert/logging"
	"github.com/x1unix/gilbert/scope"
	"time"
)

type JobContext struct {
	RootVars scope.Vars
	Logger   logging.Logger
	Context  context.Context
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

func NewJobContext(rootVars scope.Vars, log logging.Logger, ctx context.Context) JobContext {
	return JobContext{RootVars: rootVars, Logger: log}
}
