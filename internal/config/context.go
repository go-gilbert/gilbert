package config

import (
	"context"
	"os"
	"os/signal"

	"github.com/go-gilbert/gilbert/pkg/containers"
)

var ctxCell = containers.NewOnceCell(func() ctxPair {
	rootCtx := context.Background()
	ctx, cancelFn := signal.NotifyContext(rootCtx, os.Interrupt, os.Kill)
	return ctxPair{
		ctx:      ctx,
		cancelFn: cancelFn,
	}
})

type ctxPair struct {
	ctx      context.Context
	cancelFn context.CancelFunc
}

// ApplicationContext returns global application execution context.
//
// Context is terminated when SIGINT or os.Interrupt signal is received.
func ApplicationContext() context.Context {
	return ctxCell.Get().ctx
}

// CancelApplicationContext cancels application context returned from ApplicationContext().
func CancelApplicationContext() {
	ctxCell.Get().cancelFn()
}
