package manifest

import (
	"testing"

	"github.com/go-gilbert/gilbert/internal/manifest/expr"
	"github.com/stretchr/testify/assert"
)

const testFile = "./testdata/a.yaml"

func TestLoadManifest(t *testing.T) {
	expected := Manifest{
		Parser:   expr.SpecV2Parser{},
		Version:  "2",
		location: "./testdata/a.yaml",
		Imports: []string{
			"./include/b.yaml",
			"./include/c.yaml",
		},
		Vars: Vars{
			"b": "b0",
		},
		Mixins: Mixins{
			"b11mx": Mixin{
				Job{ActionName: "build"},
			},
		},
		Tasks: TaskSet{
			"build": Task{
				Job{ActionName: "build"},
			},
			"b": Task{
				Job{ActionName: "shell"},
			},
			"b1": Task{
				Job{ActionName: "shell"},
			},
			"b2": Task{
				Job{ActionName: "shell"},
			},
			"b11": Task{
				Job{ActionName: "shell"},
			},
			"c": Task{
				Job{ActionName: "shell"},
			},
		},
	}

	result, err := LoadManifest(testFile)
	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, expected, *result)
	}
}
