package yamlloader

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/hashicorp/go-set/v3"
)

type visitState uint

const (
	visitStateUnvisited visitState = iota
	visitStateVisiting
	visitStateVisited
)

const visitBufSize = 10

type LoadResult struct {
	File        manifest.JobFile
	Diagnostics parsetypes.Diagnostics
}

type LoaderConfig struct {
	BuiltinNamespaces []string
	BuiltinFlags      []string
}

type Loader struct {
	dst               manifest.JobFile
	diags             parsetypes.Diagnostics
	builtinNamespaces *set.Set[string]

	// visitState used to track cyclomatic imports.
	visitState map[string]visitState
	rootDir    string
}

func NewLoader(cfg LoaderConfig) *Loader {
	// TODO: use BuiltinFlags in loader
	return &Loader{
		visitState:        make(map[string]visitState, visitBufSize),
		builtinNamespaces: set.From(cfg.BuiltinNamespaces),
	}
}

func (l *Loader) Load(ctx context.Context, filePath string) (*LoadResult, error) {
	// abs path should be resolved only for a root file.
	// abs path for includes is resolved during decoding.
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	l.dst.Path = absPath
	l.rootDir = filepath.Dir(absPath)

	// Use DFS to be able to detect cycles.
	err = l.readInclude(ctx, &includeDecl{
		filePath: absPath,
	})

	result := &LoadResult{
		File:        l.dst,
		Diagnostics: l.diags,
	}

	return result, err
}

func (l *Loader) handleCyclomaticInclude(node *includeDecl) {
	relPath, err := filepath.Rel(l.rootDir, node.filePath)
	if err != nil {
		relPath = node.filePath
	}

	l.diags = append(l.diags, &parsetypes.Diagnostic{
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: node.location.FileName,
		Range:    node.location.Range,
		Offset:   node.location.Offset,
		Err:      fmt.Errorf("cyclomatic include of %q", relPath),
	})
}

func (l *Loader) readInclude(ctx context.Context, node *includeDecl) error {
	switch l.visitState[node.filePath] {
	case visitStateVisited:
		return nil
	case visitStateVisiting:
		l.handleCyclomaticInclude(node)
		return nil
	}

	l.visitState[node.filePath] = visitStateVisiting
	defer func() {
		l.visitState[node.filePath] = visitStateVisited
	}()

	ldCtx := newLoaderContext(ctx, loaderContext{
		dst:               &l.dst,
		fileDir:           filepath.Dir(node.filePath),
		filePath:          node.filePath,
		builtinNamespaces: l.builtinNamespaces,
	})

	result, diags, err := yamltree.ReadSource(ldCtx, jobFileSchema, yamltree.Source{
		FilePath: node.filePath,
	}, yamltree.WithUnknownFieldAction(yamltree.UnknownFieldActionWarn))

	if err != nil {
		// error breaks processing, break only if root isn't loaded
		if node.location == nil {
			return err
		}

		l.diags = append(l.diags, &parsetypes.Diagnostic{
			Severity: parsetypes.DiagnosticSeverityError,
			FileName: node.location.FileName,
			Range:    node.location.Range,
			Offset:   node.location.Offset,
			Err:      fmt.Errorf("cannot resolve include: %w", err),
		})

		return nil
	}

	l.diags = append(l.diags, diags...)
	for _, include := range result.includes {
		if err := l.readInclude(ctx, include); err != nil {
			return err
		}
	}

	return nil
}
