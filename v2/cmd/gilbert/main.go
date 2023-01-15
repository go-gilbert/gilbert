package main

import (
	"context"
	"github.com/go-gilbert/gilbert/v2/internal/cmd"
	"github.com/go-gilbert/gilbert/v2/internal/log"
)

func main() {
	args := cmd.ExpandedArgs()
	cfg, err := cmd.ParsePreRunFlags(args)
	if err != nil {
		log.Global().Fatal(err)
	}

	cmd.Run(context.Background(), cfg, cmd.ExpandedArgs())
}
