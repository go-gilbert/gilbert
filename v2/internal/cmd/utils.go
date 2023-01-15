package cmd

import (
	"context"
	"os"
	"os/signal"
)

// NewApplicationContext returns a new application context attached to process lifetime.
//
// Context is terminated on SIGINT or SIGKILL signals.
func NewApplicationContext(parentCtx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancelFn := signal.NotifyContext(parentCtx, os.Interrupt, os.Kill)
	return ctx, cancelFn
}

// ExpandedArgs returns expanded process args without executable name.
func ExpandedArgs() []string {
	if len(os.Args) > 0 {
		return os.Args[1:]
	}

	return nil
}
