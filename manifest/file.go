package manifest

import (
	"github.com/x1unix/gilbert/scope"
	"gopkg.in/yaml.v2"
)

const (
	FileName = "gilbert.yaml"
)

type Task []Job
type TaskSet map[string]Task
type RawParams map[string]interface{}

type Manifest struct {
	// Version is gilbert file format version
	Version string `yaml:"version"`

	// Imports is list of imported presets
	Imports []string `yaml:"imports,omitempty"`

	// Vars is a set of global variables
	Vars scope.Vars `yaml:"vars,omitempty"`

	// Tasks is a set of tasks
	Tasks TaskSet `yaml:"tasks,omitempty"`
}

func UnmarshalManifest(data []byte) (m *Manifest, err error) {
	m = &Manifest{}
	err = yaml.Unmarshal(data, m)
	return
}
