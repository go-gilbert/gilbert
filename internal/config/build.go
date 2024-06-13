package config

import (
	"os"
	"path/filepath"
	"runtime"
)

var (
	// Version is application version in semver format.
	// Provided during build time.
	Version = "dev"

	// Commit is Git commit SHA provided during build time.
	Commit = "local"

	// ProgramName is a base name of executable file.
	ProgramName = filepath.Base(os.Args[0])

	// Platform is target architecture and os name.
	Platform = runtime.GOARCH + "/" + runtime.GOOS
)

// ExpandedArgs returns program command line args without program name.
func ExpandedArgs() []string {
	if len(os.Args) > 1 {
		return os.Args[1:]
	}

	return nil
}
