// Package buildinfo provides build version and metadata.
package buildinfo

import "runtime"

var (
	Version  = "snapshot"
	Platform = runtime.GOOS + "/" + runtime.GOARCH
)
