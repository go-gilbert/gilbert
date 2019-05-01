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

	if p.Threshold > 100 || p.Threshold < 0 {
		return nil, fmt.Errorf("coverage threshold should be between 0 and 100 (got %f)", p.Threshold)
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
