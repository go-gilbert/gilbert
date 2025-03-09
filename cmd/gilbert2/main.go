package main

import (
	"os"

	"github.com/go-gilbert/gilbert/internal/v2/cmd"
)

func main() {
	exitCode := cmd.Main(os.Args)
	os.Exit(exitCode)
}
