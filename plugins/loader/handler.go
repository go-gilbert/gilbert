package loader

import (
	"net/url"
	"path/filepath"
)

var importHandlers = map[string]ImportHandler{
	"file": importLocalFile,
}

type ImportHandler func(*url.URL) (PluginFactory, string, error)

func importLocalFile(uri *url.URL) (PluginFactory, string, error) {
	libPath := filepath.Join(uri.Host, uri.Path)
	return loadLibrary(libPath)
}
