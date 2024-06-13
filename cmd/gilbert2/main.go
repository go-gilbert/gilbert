package main

import (
	"context"
	"errors"
	"os"

	"github.com/go-gilbert/gilbert/internal/cmd"
	"github.com/go-gilbert/gilbert/internal/config"
	"github.com/gookit/color"
)

func main() {
	err := runErr()
	if err == nil {
		return
	}

	// TODO: use logger
	if errors.Is(err, context.Canceled) {
		color.Warnln("Operation canceled by user")
		return
	}

	color.Errorf("Error: %s", err)
	os.Exit(1)
}

func runErr() error {
	ctx := config.ApplicationContext()
	defer config.CancelApplicationContext()

	c := cmd.NewCmdRoot()
	c.SetArgs(config.ExpandedArgs())

	return c.ExecuteContext(ctx)
}
