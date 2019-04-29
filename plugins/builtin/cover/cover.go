package cover

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-gilbert/gilbert-sdk"
	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/gilbert/manifest"
)

const coverFilePattern = "gbcover*.out"

// NewPlugin creates a new cover plugin instance
func NewPlugin(scope sdk.ScopeAccessor, params sdk.PluginParams, log sdk.Logger) (sdk.Plugin, error) {
	p := newParams()
	if err := mapstructure.Decode(params, &p); err != nil {
		return nil, manifest.NewPluginConfigError("cover", err)
	}

	if p.Threshold > 100 || p.Threshold < 0 {
		return nil, fmt.Errorf("coverage threshold should be between 0 and 100 (got %f)", p.Threshold)
	}

	f, err := ioutil.TempFile(os.TempDir(), coverFilePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create coverage temporary file: %s", err)
	}

	return &plugin{
		scope:     scope,
		params:    p,
		alive:     true,
		coverFile: f,
		log:       log,
	}, nil
}
