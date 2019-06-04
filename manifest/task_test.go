package manifest

import (
	sdk "github.com/go-gilbert/gilbert-sdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTask_AsyncJobsCount(t *testing.T) {
	tsk := Task{
		{Async: true},
		{Async: false},
		{Async: true},
	}

	got := tsk.AsyncJobsCount()
	assert.Equal(t, 2, got)
}

func TestTask_Clone(t *testing.T) {
	expected := Task{
		{
			Description: "foo",
			Vars: sdk.Vars{
				"v1": "foo",
				"v2": "bar",
			},
		},
	}
	origin := Task{{Description: "foo", Vars: sdk.Vars{"v1": "foo"}}}
	got := origin.Clone(sdk.Vars{"v2": "bar"})
	assert.Equal(t, expected, got)
}
