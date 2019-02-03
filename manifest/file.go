package manifest

import (
	"github.com/x1unix/gilbert/scope"
	"gopkg.in/yaml.v2"
)

const (
	// FileName is default manifest filename
	FileName = "gilbert.yaml"
)

// Task is a group of jobs
type Task []Job

// TaskSet is a set of tasks declared in a manifest file
type TaskSet map[string]Task

// RawParams is raw plugin params
type RawParams map[string]interface{}

// Manifest represents manifest file (gilbert.yaml)
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

// UnmarshalManifest parses yaml contents into Manifest structure
func UnmarshalManifest(data []byte) (m *Manifest, err error) {
	m = &Manifest{}
	err = yaml.Unmarshal(data, m)
	return
}
