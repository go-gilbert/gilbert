package manifest

import (
	"github.com/stretchr/testify/assert"
	"github.com/x1unix/gilbert/scope"
	"io/ioutil"
	"testing"
)

const testFile = "./testdata/a.yaml"

func testManifest(t *testing.T) *Manifest {
	data, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatalf("test manifest file not found at %s", testFile)
		return nil
	}

	m, err := UnmarshalManifest(data)
	if err != nil {
		t.Fatalf("failed to parse manifest file:\n  %v", err)
		return nil
	}

	m.location = testFile
	return m
}

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

	m := testManifest(t)
	tr := newImportTree(m)
	err := tr.resolveImports()
	assert.NoError(t, err)
	result := tr.result()
	assert.Equal(t, expected, result)
}
