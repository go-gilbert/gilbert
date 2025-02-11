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

var jobFileSchema = Struct[yamlJobFile](
	Field[yamlJobFile, string]("version",
		String(),
		func(ctx context.Context, dst *yamlJobFile, val string) error {
			if val != supportedManifestVersion {
				return fmt.Errorf("unsupported version: %s", val)
			}

			dst.version = val
			return nil
		},
	).Required(),
	Field[yamlJobFile, []string]("include",
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
		func(ctx context.Context, dst *yamlJobFile, val []string) error {
			dst.includes = val
			return nil
		},
	),
	Field[yamlJobFile, map[string]any](
		"const",
		Map[any](AnyScalar()),
		func(_ context.Context, dst *yamlJobFile, val map[string]any) error {
			if dst.result.Consts == nil {
				dst.result.Consts = val
			}

			copyMapUniq(dst.result.Consts, val)
			return nil
		},
	),
	Field[yamlJobFile, map[string]manifest2.InputDefinition](
		"inputs", Map[manifest2.InputDefinition](inputDefinitionSchema),
		func(_ context.Context, dst *yamlJobFile, val map[string]manifest2.InputDefinition) error {
			if len(val) == 0 {
				return errors.New("empty inputs list")
			}

			if dst.result.Inputs == nil {
				dst.result.Inputs = val
			}

			// TODO: provide previous declaration path
			return copyMapWithCheck(dst.result.Inputs, val, func(k string) error {
				return fmt.Errorf("duplicate input %q", k)
			})
		},
	),
).Constructor(func(ctx context.Context, y *yamlJobFile) error {
	c, err := getLoaderContext(ctx)
	if err != nil {
		return err
	}

	y.result = c.dst
	return nil
})
