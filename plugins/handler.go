package plugins

import (
	"context"
	"net/url"
	"path/filepath"

	"github.com/go-gilbert/gilbert/plugins/sources/github"
	"github.com/go-gilbert/gilbert/plugins/sources/gopkg"
	"github.com/go-gilbert/gilbert/plugins/sources/http"
)

// SourceProvider provides and installs plugin from source
type SourceProvider func(context.Context, *url.URL) (string, error)

var importHandlers = map[string]SourceProvider{
	"file":               getLocalPlugin,
	github.ProviderName:  github.GetPlugin,
	http.AltProviderName: http.GetPlugin,
	http.ProviderName:    http.GetPlugin,
	gopkg.ProviderName:   gopkg.GetPlugin,
}

func getLocalPlugin(_ context.Context, uri *url.URL) (string, error) {
	return filepath.Join(uri.Host, uri.Path), nil
}
