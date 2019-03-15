package manifest

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// UnmarshalManifest parses yaml contents into manifest structure
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
