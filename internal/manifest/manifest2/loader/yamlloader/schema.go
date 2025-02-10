package yamlloader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

const supportedManifestVersion = "2"

var inputDefinitionSchema = Struct(map[string]FieldVisitor[manifest2.InputDefinition]{
	"type": Field(
		Transform(String(), func(_ context.Context, _ ast.Node, s string) (manifest2.ValueType, error) {
			return manifest2.ParseValueType(s)
		}),
		func(_ context.Context, dst *manifest2.InputDefinition, val manifest2.ValueType) error {
			dst.Type = val
			return nil
		},
	).Required(),
	"format": Field(
		Transform(String(), func(_ context.Context, _ ast.Node, s string) (manifest2.ValueFormat, error) {
			return manifest2.ParseValueFormat(s)
		}),
		func(_ context.Context, dst *manifest2.InputDefinition, val manifest2.ValueFormat) error {
			dst.Format = val
			return nil
		},
	).Validation(func(_ context.Context, dst *manifest2.InputDefinition) error {
		if dst.Type != manifest2.ValueTypeString {
			return errors.New("format is available only for when type is string")
		}

		return nil
	}),
	"binding": Field(
		Pointer[manifest2.InputBinding](
			Struct(map[string]FieldVisitor[manifest2.InputBinding]{
				"env": Field(
					String(),
					func(ctx context.Context, dst *manifest2.InputBinding, val string) error {
						val = strings.TrimSpace(val)
						if val == "" {
							return errors.New("empty environment variable name")
						}
						dst.EnvVarName = val
						return nil
					},
				),
			}),
		),
		func(_ context.Context, dst *manifest2.InputDefinition, val *manifest2.InputBinding) error {
			dst.Binding = val
			return nil
		},
	),
	"items": Field(
		Pointer[manifest2.TypeSchema](
			Struct(map[string]FieldVisitor[manifest2.TypeSchema]{
				"type": Field(
					Transform(String(), func(_ context.Context, _ ast.Node, v string) (manifest2.ValueType, error) {
						return manifest2.ParseValueType(v)
					}),
					func(_ context.Context, dst *manifest2.TypeSchema, val manifest2.ValueType) error {
						if val.IsComplex() {
							return errors.New("nested complex types are not supported")
						}

						dst.Type = val
						return nil
					},
				),
				"format": Field(
					Transform(String(), func(_ context.Context, _ ast.Node, s string) (manifest2.ValueFormat, error) {
						return manifest2.ParseValueFormat(s)
					}),
					func(_ context.Context, dst *manifest2.TypeSchema, val manifest2.ValueFormat) error {
						dst.Format = val
						return nil
					},
				).Validation(func(_ context.Context, dst *manifest2.TypeSchema) error {
					if dst.Type != manifest2.ValueTypeString {
						return errors.New("format is available only for when type is string")
					}

					return nil
				}),
			}),
		),
		func(_ context.Context, dst *manifest2.InputDefinition, val *manifest2.TypeSchema) error {
			if dst.Type == manifest2.ValueTypeList {
				if val == nil {
					return errors.New("missing array element type definition")
				}
			} else if val != nil {
				return errors.New(`"items" property should be present only when "type" is "list"`)
			}

			dst.Items = val
			return nil
		},
	),
})

var jobFileSchema = Struct(
	map[string]FieldVisitor[yamlJobFile]{
		"version": Field[yamlJobFile, string](
			String(),
			func(ctx context.Context, dst *yamlJobFile, val string) error {
				if val != supportedManifestVersion {
					return fmt.Errorf("unsupported version: %s", val)
				}

				dst.version = val
				return nil
			},
		).Required(),
		"include": Field[yamlJobFile, []string](
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
		"inputs": Field[yamlJobFile, map[string]manifest2.InputDefinition](
			Map[manifest2.InputDefinition](inputDefinitionSchema),
			func(_ context.Context, dst *yamlJobFile, val map[string]manifest2.InputDefinition) error {
				if len(val) == 0 {
					return errors.New("empty inputs list")
				}

				dst.result.Inputs = val
				return nil
			},
		),
	},
).Constructor(func(ctx context.Context, y *yamlJobFile) error {
	c, err := getLoaderContext(ctx)
	if err != nil {
		return err
	}

	y.result = c.dst
	return nil
})
