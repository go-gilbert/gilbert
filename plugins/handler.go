package plugins

import (
	"context"
	"net/url"
	"path/filepath"

	"github.com/go-gilbert/gilbert/plugins/sources/http"

	"github.com/go-gilbert/gilbert/plugins/sources/github"
)

var importHandlers = map[string]SourceProvider{
	"file":               getLocalPlugin,
	github.ProviderName:  github.GetPlugin,
	http.ProviderName:    http.GetPlugin,
	http.AltProviderName: http.GetPlugin,
}

// SourceProvider provides and installs plugin from source
type SourceProvider func(context.Context, *url.URL) (string, error)

func getLocalPlugin(_ context.Context, uri *url.URL) (string, error) {
	return filepath.Join(uri.Host, uri.Path), nil
}
