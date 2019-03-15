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

func (m *Manifest) includeParent(parent *Manifest) {
	m.Vars = m.Vars.AppendNew(parent.Vars)

	if len(parent.Mixins) > 0 {
		if m.Mixins == nil {
			m.Mixins = make(Mixins)
		}

		// Copy mixins
		for k, mx := range parent.Mixins {
			// Skip if mixin with the same name defined in parent
			if _, ok := m.Mixins[k]; ok {
				continue
			}

			m.Mixins[k] = append(m.Mixins[k], mx...)
		}
	}

	if len(parent.Tasks) > 0 {
		if m.Tasks == nil {
			m.Tasks = make(TaskSet)
		}

		// Copy tasks
		for k, mx := range parent.Tasks {
			// Skip if mixin with the same name defined in parent
			if _, ok := m.Tasks[k]; ok {
				continue
			}

			m.Tasks[k] = append(m.Tasks[k], mx...)
		}
	}
}
