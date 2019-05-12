package cover

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-gilbert/gilbert-sdk"
)

const coverFilePattern = "gbcover*.out"

// NewAction creates a new cover action handler instance
func NewAction(scope sdk.ScopeAccessor, params sdk.ActionParams) (sdk.ActionHandler, error) {
	p := newParams()
	if err := params.Unmarshal(&p); err != nil {
		return nil, err
	}

	if err := p.validate(); err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile(os.TempDir(), coverFilePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create coverage temporary file: %s", err)
	}

	return &Action{
		scope:     scope,
		params:    p,
		alive:     true,
		coverFile: f,
	}, nil
}
