package cover

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParamsValidate(t *testing.T) {
	cases := map[string]struct {
		p   params
		err string
	}{
		"validate threshold below 0": {
			err: "coverage threshold should be between 0 and 100",
			p: params{
				Threshold: -1,
			},
		},
		"validate threshold above 100": {
			err: "coverage threshold should be between 0 and 100",
			p: params{
				Threshold: 101,
			},
		},
		"validate sort type": {
			err: "unsupported sort key",
			p: params{
				Threshold: 10,
				Sort:      sortParam{},
			},
		},
	}

	for n, c := range cases {
		t.Run("should "+n, func(t *testing.T) {
			err := c.p.validate()
			if c.err == "" {
				assert.NoError(t, err)
				return
			}

			if err == nil {
				t.Fatal("expected error message but got nil")
			}

			if got := err.Error(); !strings.Contains(got, c.err) {
				t.Fatalf("error '%s' should contain '%s'", got, c.err)
			}
		})
	}
}
