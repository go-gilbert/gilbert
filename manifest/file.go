package manifest

import (
	"github.com/x1unix/guru/env"
	"gopkg.in/yaml.v2"
)

const (
	FileName = "guru.yaml"
)

type Task []Job
type RawParams map[string]interface{}

type Manifest struct {
	// Version is guru file format version
	Version string `yaml:"version"`

	// Imports is list of imported presets
	Imports []string `yaml:"imports"`

	// Vars is a set of global variables
	Vars env.Vars `yaml:"vars"`

	// Tasks is a set of tasks
	Tasks map[string]Task `yaml:"tasks"`
}

func UnmarshalManifest(data []byte) (m *Manifest, err error) {
	m = &Manifest{}
	err = yaml.Unmarshal(data, m)
	return
}
