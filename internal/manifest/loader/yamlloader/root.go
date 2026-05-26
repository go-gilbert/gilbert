package yamlloader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-gilbert/gilbert/internal/manifest"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

const supportedManifestVersion = "2"

var jobFileSchema = Struct(
	Field("version",
		String(),
		func(ctx context.Context, dst *yamlJobFile, val string) error {
			if val != supportedManifestVersion {
				return fmt.Errorf("unsupported version: %s", val)
			}

			dst.version = val
			return nil
		},
	).Required(),
	Field("include",
		List(
			Transform(String(), func(ctx context.Context, n ast.Node, s string) (*includeDecl, error) {
				if s == "" {
					return nil, errors.New("empty path")
				}

				c, err := getLoaderContext(ctx)
				if err != nil {
					return nil, err
				}

				absPath := filepath.Clean(filepath.Join(c.fileDir, s))
				if absPath == c.filePath {
					return nil, errors.New("recursive include")
				}

				// drop non-existing imports.
				st, err := os.Stat(absPath)
				if err != nil {
					return nil, fmt.Errorf("cannot resolve include: %w", err)
				}

				if st.IsDir() {
					return nil, fmt.Errorf("invalid include path: %s is dir", absPath)
				}

				return &includeDecl{
					filePath: absPath,
					location: buildRefLocationWithCtx(c, n),
				}, nil
			}),
		),
		func(_ context.Context, dst *yamlJobFile, val []*includeDecl) error {
			dst.includes = val
			return nil
		},
	),
	Field(
		"plugins", importsSchema,
		func(_ context.Context, dst *yamlJobFile, val manifest.PluginImports) error {
			return dst.appendPlugins(val)
		},
	),
	Field(
		"const",
		Map(AnyScalar()),
		func(_ context.Context, dst *yamlJobFile, val map[string]any) error {
			return dst.appendConsts(val)
		},
	),
	Field(
		"inputs",
		inputsSchema,
		func(_ context.Context, dst *yamlJobFile, val manifest.Inputs) error {
			return dst.appendInputs(val)
		},
	),
	Field(
		"tasks",
		jobGroupsSchema,
		func(_ context.Context, dst *yamlJobFile, val manifest.JobGroups) error {
			return dst.appendTasks(val)
		},
	).WithContext(func(ctx context.Context) context.Context {
		return jobGroupContext(ctx, &jobGroupInfo{
			jobGroupType: manifest.JobGroupTypeTask,
		})
	}),
	Field(
		"mixins",
		jobGroupsSchema,
		func(_ context.Context, dst *yamlJobFile, val manifest.JobGroups) error {
			return dst.appendMixins(val)
		},
	).WithContext(func(ctx context.Context) context.Context {
		return jobGroupContext(ctx, &jobGroupInfo{
			jobGroupType: manifest.JobGroupTypeMixin,
		})
	}),
).Constructor(func(ctx context.Context, y *yamlJobFile) error {
	c, err := getLoaderContext(ctx)
	if err != nil {
		return err
	}

	y.result = c.dst
	y.builtinNamespaces = c.builtinNamespaces
	return nil
})
