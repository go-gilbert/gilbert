package plugins

import (
	"context"
	"github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/plugins/sources/github"
	"net/url"
	"path/filepath"
)

var importHandlers = map[string]ImportHandler{
	"file":   localFileHandler,
	"github": github.ImportHandler,
}

type ImportHandler func(context.Context, *url.URL) (sdk.PluginFactory, string, error)

func localFileHandler(_ context.Context, uri *url.URL) (sdk.PluginFactory, string, error) {
	libPath := filepath.Join(uri.Host, uri.Path)
	return loadLibrary(libPath)
}
