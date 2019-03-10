package watch

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/x1unix/gilbert/manifest"
	"github.com/x1unix/gilbert/scope"
	"path/filepath"
)

var defaultDebounceTime = manifest.Period(1000)

type params struct {
	Path         string
	Ignore       []string
	DebounceTime manifest.Period
	Job          *manifest.Job
	rootDir      string
	blacklist    []string
}

// pathIgnored checks if path matches ignore list
func (p *params) pathIgnored(path string) (bool, error) {
	for _, ignoredItem := range p.blacklist {
		match, err := filepath.Match(ignoredItem, path)
		if err != nil {
			return false, err
		}

		if match {
			return true, nil
		}
	}

	return false, nil
}

func parseParams(raw manifest.RawParams, scope *scope.Scope) (*params, error) {
	p := params{
		DebounceTime: defaultDebounceTime,
	}
	if err := mapstructure.Decode(raw, &p); err != nil {
		return nil, fmt.Errorf("failed to read configuration: %s", err)
	}

	if p.Job == nil {
		return nil, fmt.Errorf("job to run is not defined")
	}

	if err := scope.Scan(&p.Path); err != nil {
		return nil, err
	}

	if p.Path == "" {
		return nil, fmt.Errorf("watch path is undefined, please set path to watch in 'path' parameter")
	}

	if len(p.Ignore) == 0 {
		return &p, nil
	}

	// Convert paths in ignore list from relative to absolute
	p.blacklist = make([]string, len(p.Ignore))
	rootPath := scope.Environment.ProjectDirectory
	for _, relPath := range p.Ignore {
		// Expand variables in value
		relPath, err := scope.ExpandVariables(relPath)
		if err != nil {
			return nil, err
		}

		// Convert to absolute path if it's relative path
		if !filepath.IsAbs(relPath) {
			relPath = filepath.Join(rootPath, relPath)
		}

		p.blacklist = append(p.blacklist, relPath)
	}

	return &p, nil
}
