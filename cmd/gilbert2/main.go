package main

import (
	"os"

	"github.com/go-gilbert/gilbert/internal/v2/cmd"
	"github.com/go-gilbert/gilbert/internal/v2/log"
)

func main() {
	exitCode := cmd.Main(log.DefaultIOStreams, os.Args)
	os.Exit(exitCode)
}
