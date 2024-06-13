//go:build linux || darwin
// +build linux darwin

package loader

import (
	"fmt"
	"plugin"

	"github.com/go-gilbert/gilbert-sdk"
)

const (
	pluginActionsProc = "GetPluginActions"
	pluginNameProc    = "GetPluginName"
)

func badSymbolTypeErr(symName string, expected, got interface{}) error {
	return fmt.Errorf("invalid %s() symbol signature (want %T, but got %T)", symName, expected, got)
}

// LoadPlugin loads plugin from provided source
func LoadPlugin(libPath string) (pluginName string, pluginActions sdk.Actions, err error) {
	handle, err := plugin.Open(libPath)

	if err != nil {
		return "", nil, err
	}

	pluginName, err = getPluginName(handle)
	if err != nil {
		return "", nil, err
	}

	pluginActions, err = getPluginActions(handle)
	return pluginName, pluginActions, err
}

func getPluginActions(handle *plugin.Plugin) (actions sdk.Actions, err error) {
	defer func() {
		// panic handler, just in for safety
		if r := recover(); r != nil {
			err = fmt.Errorf("got panic when trying to get plugin actions: %s", r)
		}
	}()

	procHandle, err := handle.Lookup(pluginActionsProc)
	if err != nil {
		return nil, fmt.Errorf("cannot get plugin factory (%s)", err)
	}

	fn, ok := procHandle.(func() sdk.Actions)
	if !ok {
		return nil, badSymbolTypeErr(pluginActionsProc, fn, procHandle)
	}

	return fn(), nil
}

func getPluginName(handle *plugin.Plugin) (string, error) {
	procHandle, err := handle.Lookup(pluginNameProc)
	if err != nil {
		return "", fmt.Errorf("cannot get plugin name (%s)", err)
	}

	nameFn, ok := procHandle.(func() string)
	if !ok {
		return "", badSymbolTypeErr(pluginNameProc, nameFn, procHandle)
	}

	return nameFn(), nil
}
