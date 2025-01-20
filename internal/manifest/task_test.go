package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			Vars: Vars{
				"v1": "foo",
				"v2": "bar",
			},
		},
	}
	origin := Task{{Description: "foo", Vars: Vars{"v1": "foo"}}}
	got := origin.Clone(Vars{"v2": "bar"})
	assert.Equal(t, expected, got)
}
