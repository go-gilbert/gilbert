//go:build !linux && !darwin
// +build !linux,!darwin

package loader

import (
	"errors"

	"github.com/go-gilbert/gilbert-sdk"
)

// LoadPlugin loads plugin from provided source
func LoadPlugin(libPath string) (pluginName string, pluginActions sdk.Actions, err error) {
	return "", nil, errors.New("plugins currently are not supported on this platform")
}
