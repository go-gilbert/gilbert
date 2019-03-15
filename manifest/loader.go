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

// LoadManifest loads manifest from specified path and it's imports
func LoadManifest(path string) (*Manifest, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("manifest file not found at '%s'", path)
	}

	m, err := UnmarshalManifest(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest file:\n  %v", err)
	}

	m.location = path

	// Return as-is if no imports declared
	if len(m.Imports) == 0 {
		return m, nil
	}

	// Load imports
	tree := newImportTree(m)
	if err := tree.resolveImports(); err != nil {
		return nil, fmt.Errorf("failed to resolve imports in manifest file - %s", err)
	}

	result := tree.result()
	return &result, nil
}

// FromDirectory loads gilbert.yaml from specified directory
func FromDirectory(dir string) (m *Manifest, err error) {
	location := filepath.Join(dir, FileName)
	return LoadManifest(location)
}
