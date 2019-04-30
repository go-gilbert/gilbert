package plugins

import (
	"context"
	"fmt"
	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/plugins/builtin"
	"github.com/go-gilbert/gilbert/plugins/loader"
	"net/url"
)

var registry = make(map[string]sdk.PluginFactory)

func Import(ctx context.Context, pluginUrl string) error {
	if err := registerPluginFromUrl(ctx, pluginUrl); err != nil {
		return fmt.Errorf("failed to load plugin from '%s':\n%s", pluginUrl, err)
	}

	return nil
}

func Get(pluginName string) (sdk.PluginFactory, error) {
	if plug, ok := registry[pluginName]; ok {
		return plug, nil
	}

	if plug, ok := builtin.DefaultPlugins[pluginName]; ok {
		return plug, nil
	}

	return nil, fmt.Errorf("plugin '%s' not found", pluginName)
}

// Loaded checks if plugin is already loaded
func Loaded(name string) bool {
	_, ok := registry[name]
	return ok
}

func registerPluginFromUrl(ctx context.Context, pluginUrl string) error {
	uri, err := url.Parse(pluginUrl)
	if err != nil {
		return fmt.Errorf("invalid plugin import URL (%s)", err)
	}

	if uri.Scheme == "" {
		return fmt.Errorf("invalid plugin import URL")
	}

	importHandler, ok := importHandlers[uri.Scheme]
	if !ok {
		return fmt.Errorf("unsupported plugin URL handler: '%s'", uri.Scheme)
	}

	pluginPath, err := importHandler(ctx, uri)
	if err != nil {
		return fmt.Errorf("failed to import plugin: %s", err)
	}

	pf, pName, err := loader.LoadLibrary(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %s", err)
	}

	if Loaded(pName) {
		return fmt.Errorf("plugin '%s' is already loaded", pName)
	}

	log.Default.Debugf("loaded plugin '%s' from '%s'", pName, pluginPath)
	registry[pName] = pf
	return nil
}
