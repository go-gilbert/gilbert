package plugins

import (
	"context"
	"fmt"
	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/plugins/builtin"
	"net/url"
)

var registry = make(map[string]sdk.PluginFactory)

func Import(ctx context.Context, pluginUrl string) error {
	if err := registerPluginFromUrl(ctx, pluginUrl); err != nil {
		return fmt.Errorf("failed to load plugin '%s':\n%s", pluginUrl, err)
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

	pf, pName, err := importHandler(ctx, uri)
	if err != nil {
		return fmt.Errorf("failed to import plugin: %s", err)
	}

	log.Default.Debugf("loaded plugin '%s' from '%s'", pName, pluginUrl)
	registry[pName] = pf
	return nil
}
