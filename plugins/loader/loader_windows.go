package loader

import (
	"errors"

	"github.com/go-gilbert/gilbert-sdk"
)

// LoadLibrary loads library from provided source
func LoadLibrary(libPath string) (sdk.PluginFactory, string, error) {
	return nil, "", errors.New("plugins currently are not supported on Windows")
}
