package main

import (
	"os"

	"github.com/go-gilbert/gilbert/internal/cmd"
	"github.com/go-gilbert/gilbert/internal/log"
)

func main() {
	exitCode := cmd.Main(log.DefaultIOStreams, os.Args)
	os.Exit(exitCode)
}
