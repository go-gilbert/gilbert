package plugins

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/actions"
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/plugins/loader"
)

func formatPluginActionName(pluginName, actionName string) string {
	return pluginName + ":" + actionName
}

func registerPluginAction(pName, hName string, handler sdk.HandlerFactory) error {
	hName = strings.TrimSpace(hName)
	actionName := formatPluginActionName(pName, hName)
	if err := actions.HandleFunc(actionName, handler); err != nil {
		return err
	}

	log.Default.Debugf("loader: registered action handler '%s'", hName)
	return nil
}

// Import imports plugin from URL and loads it
func Import(ctx context.Context, pluginUrl string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to load plugin from '%s'\n:%s", pluginUrl, err)
		}
	}()

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

	pluginName, handlers, err := loader.LoadPlugin(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %s", err)
	}

	pluginName = strings.TrimSpace(pluginName)
	if pluginName == "" {
		return errors.New("plugin name should not be empty")
	}

	log.Default.Debugf("loader: loaded plugin '%s' from '%s'", pluginName, pluginPath)

	// register plugin action handlers
	for hName, handler := range handlers {
		if err := registerPluginAction(pluginName, hName, handler); err != nil {
			return fmt.Errorf("failed to register action handler, %s", err)
		}
	}
	return nil
}
