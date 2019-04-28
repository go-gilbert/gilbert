// +build !windows,!js,!nacl

package loader

import (
	"fmt"
	"github.com/x1unix/gilbert/log"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
	"plugin"

	"github.com/x1unix/gilbert/plugins"
)

const (
	newPluginProc  = "NewPlugin"
	pluginNameProc = "GetPluginName"
)

func badSymbolTypeErr(symName string, got, want interface{}) error {
	return fmt.Errorf("invalid %s() symbol signature (want %T, but got %T)", want, got)
}

func loadLibrary(libPath string) (plugins.PluginFactory, string, error) {
	handle, err := plugin.Open(libPath)

	if err != nil {
		return nil, "", fmt.Errorf("failed to load plugin, %s (file '%s')", err, libPath)
	}

	name, err := getPluginName(handle)
	if err != nil {
		return nil, "", err
	}

	factory, err := getPluginFactory(handle)
	if err != nil {
		return nil, "", err
	}

	return factory, name, nil
}

func getPluginFactory(handle *plugin.Plugin) (plugins.PluginFactory, error) {
	procHandle, err := handle.Lookup(newPluginProc)
	if err != nil {
		return nil, fmt.Errorf("cannot get plugin factory (%s)", err)
	}

	fn, ok := procHandle.(func(*scope.Scope, manifest.RawParams, log.Logger) (plugins.Plugin, error))
	if !ok {
		return nil, badSymbolTypeErr(pluginNameProc, procHandle, fn)
	}

	return fn, nil
}

func getPluginName(handle *plugin.Plugin) (string, error) {
	procHandle, err := handle.Lookup(pluginNameProc)
	if err != nil {
		return "", fmt.Errorf("cannot get plugin name (%s)", err)
	}

	nameFn, ok := procHandle.(func() string)
	if !ok {
		return "", badSymbolTypeErr(pluginNameProc, procHandle, nameFn())
	}

	return nameFn(), nil
}
