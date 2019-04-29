package manifest

import (
	"github.com/go-gilbert/gilbert-sdk"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testFile = "./testdata/a.yaml"

func TestLoadManifest(t *testing.T) {
	expected := Manifest{
		Version:  "1.0",
		location: "./testdata/a.yaml",
		Imports: []string{
			"./include/b.yaml",
			"./include/c.yaml",
		},
		Vars: sdk.Vars{
			"b": "b0",
		},
		Mixins: Mixins{
			"b11mx": Mixin{
				sdk.Job{PluginName: "build"},
			},
		},
		Tasks: TaskSet{
			"build": Task{
				sdk.Job{PluginName: "build"},
			},
			"b": Task{
				sdk.Job{PluginName: "shell"},
			},
			"b1": Task{
				sdk.Job{PluginName: "shell"},
			},
			"b2": Task{
				sdk.Job{PluginName: "shell"},
			},
			"b11": Task{
				sdk.Job{PluginName: "shell"},
			},
			"c": Task{
				sdk.Job{PluginName: "shell"},
			},
		},
	}

	result, err := LoadManifest(testFile)
	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, expected, *result)
	}
}
