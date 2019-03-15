package manifest

import (
	"github.com/stretchr/testify/assert"
	"github.com/x1unix/gilbert/scope"
	"testing"
)

const testFile = "./testdata/a.yaml"

func TestImportResolver_BuildTree(t *testing.T) {
	expected := Manifest{
		Version:  "1.0",
		location: "./testdata/a.yaml",
		Imports: []string{
			"./include/b.yaml",
			"./include/c.yaml",
		},
		Vars: scope.Vars{
			"b": "b0",
		},
		Mixins: Mixins{
			"b11mx": Mixin{
				Job{PluginName: "build"},
			},
		},
		Tasks: TaskSet{
			"b": Task{
				Job{PluginName: "shell"},
			},
			"b1": Task{
				Job{PluginName: "shell"},
			},
			"b2": Task{
				Job{PluginName: "shell"},
			},
			"b11": Task{
				Job{PluginName: "shell"},
			},
			"c": Task{
				Job{PluginName: "shell"},
			},
		},
	}

	result, err := LoadManifest(testFile)
	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, expected, *result)
	}
}
