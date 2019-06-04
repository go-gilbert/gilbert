package manifest

import (
	sdk "github.com/go-gilbert/gilbert-sdk"
	"github.com/stretchr/testify/assert"
	"testing"
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

	vars := sdk.Vars{
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
