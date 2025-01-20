//go:build linux || darwin
// +build linux darwin

package loader

import (
	"errors"

	"github.com/go-gilbert/gilbert/internal/runner"
)

// LoadPlugin loads plugin from provided source
func LoadPlugin(libPath string) (pluginName string, pluginActions runner.ActionHandlers, err error) {
	// TODO: reimplement
	return "", nil, errors.New("plugin feature not implemented")
}
