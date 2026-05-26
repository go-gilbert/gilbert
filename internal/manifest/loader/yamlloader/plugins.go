package yamlloader

import (
	"context"
	"fmt"

	"github.com/go-gilbert/gilbert/internal/manifest"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

var importsSchema = Map(
	Transform(String(), func(ctx context.Context, n ast.Node, s string) (*manifest.PluginImport, error) {
		loc, err := buildRefLocation(ctx, n)
		if err != nil {
			return nil, err
		}

		c, err := getLoaderContext(ctx)
		if err != nil {
			return nil, err
		}

		importURI, err := manifest.PathIntoURI(s, c.fileDir)
		if err != nil {
			return nil, fmt.Errorf("cannot parse plugin URL: %w", err)
		}

		return &manifest.PluginImport{
			Location: loc,
			URI:      importURI,
		}, nil
	}),
)
