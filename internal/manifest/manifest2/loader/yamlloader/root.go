package yamlloader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
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
			Transform(String(), func(ctx context.Context, _ ast.Node, s string) (string, error) {
				if s == "" {
					return "", errors.New("empty path")
				}

				c, err := getLoaderContext(ctx)
				if err != nil {
					return "", err
				}

				absPath := filepath.Clean(filepath.Join(c.fileDir, s))
				if absPath == c.filePath {
					return "", errors.New("recursive include")
				}

				// drop non-existing imports.
				st, err := os.Stat(absPath)
				if err != nil {
					return "", fmt.Errorf("cannot resolve include: %w", err)
				}

				if st.IsDir() {
					return "", fmt.Errorf("invalid include path: %s is dir", absPath)
				}

				return absPath, nil
			}),
		),
		func(_ context.Context, dst *yamlJobFile, val []string) error {
			dst.includes = val
			return nil
		},
	),
	Field(
		"plugins", importsSchema,
		func(_ context.Context, dst *yamlJobFile, val manifest2.PluginImports) error {
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
		func(_ context.Context, dst *yamlJobFile, val manifest2.Inputs) error {
			return dst.appendInputs(val)
		},
	),
	Field(
		"tasks",
		jobGroupsSchema,
		func(_ context.Context, dst *yamlJobFile, val manifest2.JobGroups) error {
			return dst.appendTasks(val)
		},
	).WithContext(func(ctx context.Context) context.Context {
		return jobGroupContext(ctx, &jobGroupInfo{
			jobGroupType: manifest2.JobGroupTypeTask,
		})
	}),
	Field(
		"mixins",
		jobGroupsSchema,
		func(_ context.Context, dst *yamlJobFile, val manifest2.JobGroups) error {
			return dst.appendMixins(val)
		},
	).WithContext(func(ctx context.Context) context.Context {
		return jobGroupContext(ctx, &jobGroupInfo{
			jobGroupType: manifest2.JobGroupTypeMixin,
		})
	}),
).Constructor(func(ctx context.Context, y *yamlJobFile) error {
	c, err := getLoaderContext(ctx)
	if err != nil {
		return err
	}

	y.result = c.dst
	y.knownNamespaces = c.knownNamespaces
	return nil
})
