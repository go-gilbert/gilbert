package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManifest_Location(t *testing.T) {
	want := "./foo"
	m := Manifest{location: want}
	assert.Equal(t, want, m.Location())
}

func TestMixin_ToTask(t *testing.T) {
	m := Mixin{
		{ActionName: "build", Async: true},
		{ActionName: "shell", Async: true},
	}

	vars := Vars{
		"foo": "bar",
		"bar": "foo",
	}

	expected := Task{
		{ActionName: "build", Async: true, Vars: vars},
		{ActionName: "shell", Async: true, Vars: vars},
	}

	got := m.ToTask(vars)
	assert.Equal(t, expected, got)
}
