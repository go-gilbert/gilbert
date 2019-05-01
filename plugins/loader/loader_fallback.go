// +build !linux,!darwin

package loader

import (
	"context"
	"fmt"
	"github.com/go-gilbert/gilbert-sdk"
)

// LoadPlugin loads plugin from provided source
func LoadPlugin(ctx context.Context, libPath string) (pluginName string, pluginActions sdk.Actions, err error) {

	bridge, err := newPluginBridge(ctx, libPath)
	if err != nil {
		return "", nil, fmt.Errorf("cannot get plugin name: %s", err)
	}

	defer func() {
		if err != nil {
			bridge.Dispose()
		}
	}()

	pluginName, err = bridge.PluginName()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get plugin name: %s", err)
	}

	pluginActions, err = bridge.PluginActions()
	if err != nil {
		return pluginName, nil, fmt.Errorf("failed to get plugin actions: %s", err)
	}

	loadedPlugins[pluginName] = bridge
	return pluginName, pluginActions, nil
}
