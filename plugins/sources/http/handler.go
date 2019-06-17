package http

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/plugins/support"
	"github.com/go-gilbert/gilbert/storage"
	"github.com/go-gilbert/gilbert/support/fs"
	"github.com/go-gilbert/gilbert/support/web"
)

const (
	// ProviderName is plugin provider name
	ProviderName = "http"

	// AltProviderName is alternative provider name
	AltProviderName = "https"

	defaultPluginFName = "plugin"
)

func getPluginDirectory(uri string) string {
	hasher := md5.New()
	// nolint:errcheck
	hasher.Write([]byte(uri))
	return filepath.Join(ProviderName, hex.EncodeToString(hasher.Sum(nil)))
}

// GetPlugin is web source handler for plugins
func GetPlugin(ctx context.Context, uri *url.URL) (string, error) {
	strURL := uri.String()
	dir, err := storage.Path(storage.Plugins, getPluginDirectory(strURL))
	if err != nil {
		return "", err
	}

	// TODO: determine real file name from web response
	pluginPath := filepath.Join(dir, support.AddPluginExtension(defaultPluginFName))
	exists, err := fs.Exists(pluginPath)
	if err != nil {
		return "", err
	}

	if !exists {
		log.Default.Debugf("http: init plugin directory: '%s'", dir)
		if err = os.MkdirAll(dir, support.PluginPermissions); err != nil {
			return "", err
		}

		log.Default.Logf("Downloading plugin file from '%s'...", strURL)
		if err := web.ProgressDownloadFile(&http.Client{}, strURL, pluginPath); err != nil {
			return "", err
		}
	}

	return pluginPath, nil
}
