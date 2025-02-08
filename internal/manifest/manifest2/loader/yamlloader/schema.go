package yamlloader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/go-gilbert/gilbert/pkg/yamltree"
)

var jobFileSchema = Struct[yamlJobFile](
	map[string]FieldVisitor[yamlJobFile]{
		"version": Field[yamlJobFile, string](
			true, String(),
			func(ctx context.Context, dst *yamlJobFile, val string) error {
				if val != "2" {
					return fmt.Errorf("invalid version: %s", val)
				}

				dst.version = val
				return nil
			},
		),
		"include": Field[yamlJobFile, []string](
			false,
			List[string](
				Transform[string](String(), func(ctx context.Context, s string) (string, error) {
					c, err := getLoaderContext(ctx)
					if err != nil {
						return "", err
					}

					absPath := filepath.Clean(filepath.Join(c.fileDir, s))
					_, err = os.Stat(absPath)
					if err != nil {
						return s, fmt.Errorf("cannot resolve import: %w", err)
					}

					return absPath, nil
				}),
			),
			func(ctx context.Context, dst *yamlJobFile, val []string) error {
				dst.includes = val
				return nil
			},
		),
	},
)
