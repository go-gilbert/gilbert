package manifest

import (
	"github.com/x1unix/gilbert/scope"
)

const (
	// FileName is default manifest filename
	FileName = "gilbert.yaml"
)

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

	// Mixins is a set of declared mixins
	Mixins Mixins `yaml:"mixins,omitempty"`

	// location is manifest location
	location string `yaml:"-"`
}

// Location returns manifest file location, if it was loaded using FromDirectory method
func (m *Manifest) Location() string {
	return m.location
}
