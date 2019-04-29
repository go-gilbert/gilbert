package loader

import (
	"github.com/go-gilbert/gilbert-sdk"
	"net/url"
	"path/filepath"
)

var importHandlers = map[string]ImportHandler{
	"file": importLocalFile,
}

type ImportHandler func(*url.URL) (sdk.PluginFactory, string, error)

func importLocalFile(uri *url.URL) (sdk.PluginFactory, string, error) {
	libPath := filepath.Join(uri.Host, uri.Path)
	return loadLibrary(libPath)
}
