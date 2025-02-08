package yamlloader

import (
	"context"
	"fmt"

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
					// TODO: check & transform
					return s, nil
				}),
			),
			func(ctx context.Context, dst *yamlJobFile, val []string) error {
				dst.includes = val
				return nil
			},
		),
	},
)
