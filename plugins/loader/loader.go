// +build !windows,!js,!nacl

package loader

import (
	"fmt"
	"github.com/go-gilbert/gilbert-sdk"
	"plugin"
)

const (
	newPluginProc  = "NewPlugin"
	pluginNameProc = "GetPluginName"
)

func badSymbolTypeErr(symName string, got, want interface{}) error {
	return fmt.Errorf("invalid %s() symbol signature (want %T, but got %T)", symName, want, got)
}

func loadLibrary(libPath string) (sdk.PluginFactory, string, error) {
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

func getPluginFactory(handle *plugin.Plugin) (sdk.PluginFactory, error) {
	procHandle, err := handle.Lookup(newPluginProc)
	if err != nil {
		return nil, fmt.Errorf("cannot get plugin factory (%s)", err)
	}

	fn, ok := procHandle.(func(sdk.ScopeAccessor, sdk.PluginParams, sdk.Logger) (sdk.Plugin, error))
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
