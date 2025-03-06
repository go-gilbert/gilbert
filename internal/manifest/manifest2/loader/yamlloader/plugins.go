package yamlloader

import (
	"context"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

var importsSchema = Map(
	Transform(String(), func(ctx context.Context, n ast.Node, s string) (*manifest2.PluginImport, error) {
		loc, err := buildRefLocation(ctx, n)
		if err != nil {
			return nil, err
		}

		return &manifest2.PluginImport{
			Location: loc,
			URL:      s,
		}, nil
	}),
)
