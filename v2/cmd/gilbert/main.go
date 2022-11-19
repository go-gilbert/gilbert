package main

import (
	"context"
	"os"
	"os/signal"
)

func main() {
	ctx, cancelFn := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

}
