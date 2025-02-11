package yamlloader

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

var listTypeSchema = Struct[manifest2.TypeSchema](
	Field("type",
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
	Field("format",
		Transform(String(), func(_ context.Context, _ ast.Node, s string) (manifest2.ValueFormat, error) {
			return manifest2.ParseValueFormat(s)
		}),
		func(_ context.Context, dst *manifest2.TypeSchema, val manifest2.ValueFormat) error {
			dst.Format = val
			return nil
		},
	).Validation(func(_ context.Context, dst *manifest2.TypeSchema) error {
		if dst.Type != manifest2.ValueTypeString {
			return errors.New("format is available only when type is string")
		}

		return nil
	}),
	Field("dateFormat", String(),
		func(_ context.Context, dst *manifest2.TypeSchema, s string) error {
			if dst.Format != manifest2.ValueFormatDate {
				return errors.New(`"dateFormat" can specified only when "format" field set to "date"`)
			}

			if s == "" {
				return errors.New("empty format")
			}

			dst.DateFormat = s
			return nil
		},
	),
)

var inputDefinitionSchema = Struct[manifest2.InputDefinition](
	Field("type",
		Transform(String(), func(_ context.Context, _ ast.Node, s string) (manifest2.ValueType, error) {
			return manifest2.ParseValueType(s)
		}),
		func(_ context.Context, dst *manifest2.InputDefinition, val manifest2.ValueType) error {
			dst.Type = val
			return nil
		},
	).Required(),
	Field("format",
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
	Field("dateFormat", String(),
		func(_ context.Context, dst *manifest2.InputDefinition, s string) error {
			if dst.Format != manifest2.ValueFormatDate {
				return errors.New(`"dateFormat" can specified only when "format" field set to "date"`)
			}

			if s == "" {
				return errors.New("empty format")
			}

			dst.DateFormat = s
			return nil
		},
	),
	FieldFunc("default",
		getDefaultValueVisitor,
		func(ctx context.Context, dst *manifest2.InputDefinition, val any) error {
			dst.DefaultValue = val
			return nil
		},
	),
	Field("binding",
		Pointer[manifest2.InputBinding](
			Struct[manifest2.InputBinding](
				Field("env", String(),
					func(ctx context.Context, dst *manifest2.InputBinding, val string) error {
						val = strings.TrimSpace(val)
						if val == "" {
							return errors.New("empty environment variable name")
						}
						dst.EnvVarName = val
						return nil
					},
				),
			),
		),
		func(_ context.Context, dst *manifest2.InputDefinition, val *manifest2.InputBinding) error {
			dst.Binding = val
			return nil
		},
	),
	Field("items",
		Pointer[manifest2.TypeSchema](listTypeSchema),
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
).Validation(func(ctx context.Context, n ast.Node, dst *manifest2.InputDefinition) error {
	c, err := getLoaderContext(ctx)
	if err != nil {
		return err
	}

	rng, offset := GetNodeRange(n)
	dst.Location = manifest2.ReferenceLocation{
		FileName: c.filePath,
		Range:    rng,
		Offset:   offset,
	}

	return nil
})

func getDefaultValueVisitor(_ context.Context, node ast.Node, def *manifest2.InputDefinition) (ValueVisitor[any], error) {
	if !IsPrimitiveNode(node) {
		return nil, fmt.Errorf("expected %s but got %s", def.Type, node.Type())
	}

	switch def.Type {
	case manifest2.ValueTypeString:
		return Transform[string, any](String(), func(ctx context.Context, node ast.Node, s string) (any, error) {
			switch def.Format {
			case manifest2.ValueFormatDate:
				format := def.DateFormat
				if format == "" {
					format = manifest2.DefaultDateFormat
				}

				return time.Parse(format, s)
			case manifest2.ValueFormatDuration:
				return time.ParseDuration(s)
			case manifest2.ValueFormatURL:
				_, err := url.Parse(s)
				return s, err
			}
			return s, nil
		}), nil
	case manifest2.ValueTypeInt:
		return Transform[int64, any](Int[int64](), intoAny), nil
	case manifest2.ValueTypeFloat:
		return Transform[float64, any](Float[float64](), intoAny), nil
	case manifest2.ValueTypeBool:
		return Transform[bool, any](Bool(), intoAny), nil
	default:
		// TODO: support lists?
		return nil, fmt.Errorf("default values for input of type %q are not supported", def.Type)
	}
}

func intoAny[T any](_ context.Context, _ ast.Node, s T) (any, error) {
	return s, nil
}
