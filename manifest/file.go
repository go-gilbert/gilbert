package manifest

import (
	"fmt"
	"github.com/x1unix/gilbert/scope"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
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

	// Mixins is a set of declared mixins
	Mixins Mixins `yaml:"mixins,omitempty"`

	// location is manifest location
	location string `yaml:"-"`
}

// Location returns manifest file location, if it was loaded using FromDirectory method
func (m *Manifest) Location() string {
	return m.location
}

// UnmarshalManifest parses yaml contents into Manifest structure
func UnmarshalManifest(data []byte) (m *Manifest, err error) {
	m = &Manifest{}
	err = yaml.Unmarshal(data, m)
	return
}

// FromDirectory loads gilbert.yaml from specified directory
func FromDirectory(dir string) (m *Manifest, err error) {
	location := filepath.Join(dir, FileName)
	data, err := ioutil.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("manifest file not found (%s) at %s", FileName, dir)
	}

	m, err = UnmarshalManifest(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest file:\n  %v", err)
	}

	m.location = location
	return m, nil
}
