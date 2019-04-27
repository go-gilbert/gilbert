package loader

import (
	"github.com/x1unix/gilbert/plugins"
	"net/url"
	"path/filepath"
)

var importHandlers = map[string]ImportHandler{
	"file": importLocalFile,
}

type ImportHandler func(*url.URL) (plugins.PluginFactory, string, error)

func importLocalFile(uri *url.URL) (plugins.PluginFactory, string, error) {
	libPath := filepath.Join(uri.Host, uri.Path)
	return loadLibrary(libPath)
}
