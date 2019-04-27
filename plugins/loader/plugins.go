package loader

import (
	"fmt"
	"github.com/x1unix/gilbert/log"
	"github.com/x1unix/gilbert/plugins"
	"github.com/x1unix/gilbert/plugins/builtin"
	"net/url"
)

var registry = make(map[string]plugins.PluginFactory)

func Import(pluginUrl string) error {
	if err := registerPluginFromUrl(pluginUrl); err != nil {
		return fmt.Errorf("failed to load plugin '%s', %s", pluginUrl, err)
	}

	log.Default.Debugf("found plugin '%s'", pluginUrl)
	return nil
}

func Get(pluginName string) (plugins.PluginFactory, error) {
	if plug, ok := registry[pluginName]; ok {
		return plug, nil
	}

	if plug, ok := builtin.DefaultPlugins[pluginName]; ok {
		return plug, nil
	}

	return nil, fmt.Errorf("plugin '%s' not found", pluginName)
}

func registerPluginFromUrl(pluginUrl string) error {
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

	pf, pName, err := importHandler(uri)
	if err != nil {
		return fmt.Errorf("failed to import plugin: %s", err)
	}

	registry[pName] = pf
	return nil
}
