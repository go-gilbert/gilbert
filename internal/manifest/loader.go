package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-gilbert/gilbert/internal/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/manifest/template"
	"github.com/goccy/go-yaml"
)

// UnmarshalManifest parses yaml contents into manifest structure
func UnmarshalManifest(data []byte) (m *Manifest, err error) {
	parsed, err := template.CompileManifest(data)
	if err != nil {
		return nil, fmt.Errorf("template syntax error in manifest file: %s", err)
	}

	m = &Manifest{}
	if err := yaml.Unmarshal(parsed, m); err != nil {
		// Return formatted error
		return nil, fmt.Errorf("%s\n\n[ExpressionError in file]:\n%s", err, string(parsed))
	}
	return
}

// LoadManifest loads manifest from specified path and it's imports
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("manifest file not found at %q", path)
	}

	m, err := UnmarshalManifest(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest file:\n  %w", err)
	}

	exprParser, err := expr.GetParser(m.Version)
	if err != nil {
		return nil, err
	}

	m.location = path

	// Return as-is if no imports declared
	if len(m.Imports) == 0 {
		return m, nil
	}

	// Load imports
	tree := newImportTree(m)
	if err := tree.resolveImports(); err != nil {
		return nil, fmt.Errorf("failed to resolve imports in manifest file - %w", err)
	}

	result := tree.result()
	result.Parser = exprParser
	return &result, nil
}

// FromDirectory loads gilbert.yaml from specified directory
func FromDirectory(dir string) (m *Manifest, err error) {
	location := filepath.Join(dir, FileName)
	return LoadManifest(location)
}
