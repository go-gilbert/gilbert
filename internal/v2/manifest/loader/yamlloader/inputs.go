package yamlloader

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

var listTypeSchema = Struct(
	Field("type",
		Transform(String(), func(_ context.Context, _ ast.Node, v string) (manifest.ValueType, error) {
			return manifest.ParseValueType(v)
		}),
		func(_ context.Context, dst *manifest.TypeSchema, val manifest.ValueType) error {
			if val.IsComplex() {
				return errors.New("nested complex types are not supported")
			}

			dst.Type = val
			return nil
		},
	),
	Field("dateFormat", String(),
		func(_ context.Context, dst *manifest.TypeSchema, s string) error {
			if dst.Type != manifest.ValueTypeDate {
				return errors.New(`"dateFormat" can specified only when "type" field set to "date"`)
			}

			if s == "" {
				return errors.New("empty format")
			}

			dst.DateFormat = s
			return nil
		},
	),
)

var inputsSchema = Map(Pointer(inputDefinitionSchema)).
	CollectDoc(func(_ context.Context, fi FieldInfo, def *manifest.InputDefinition) *manifest.InputDefinition {
		def.Name = fi.Key
		def.Doc = fi.Doc
		return def
	})

var inputDefinitionSchema = Struct(
	Field("type",
		Transform(String(), func(_ context.Context, _ ast.Node, s string) (manifest.ValueType, error) {
			return manifest.ParseValueType(s)
		}),
		func(_ context.Context, dst *manifest.InputDefinition, val manifest.ValueType) error {
			dst.Schema.Type = val
			return nil
		},
	).Required(),
	Field("dateFormat", String(),
		func(_ context.Context, dst *manifest.InputDefinition, s string) error {
			if dst.Schema.Type != manifest.ValueTypeDate {
				return errors.New(`"dateFormat" can specified only when "type" field set to "date"`)
			}

			if s == "" {
				return errors.New("empty format")
			}

			dst.Schema.DateFormat = s
			return nil
		},
	),
	FieldFunc("default",
		func(_ context.Context, _ ast.Node, dst *manifest.InputDefinition) (ValueVisitor[*manifest.TypedLazyValue], error) {
			return newDefaultValVisitor(dst), nil
		},
		func(ctx context.Context, dst *manifest.InputDefinition, val *manifest.TypedLazyValue) error {
			dst.DefaultValue = val
			return nil
		},
	),
	Field("binding",
		Pointer(
			Struct(
				Field("env", String(),
					func(_ context.Context, dst *manifest.InputBinding, val string) error {
						val = strings.TrimSpace(val)
						if val == "" {
							return errors.New("empty environment variable name")
						}
						dst.EnvVarName = val
						return nil
					},
				),
				Field("flag", String(),
					func(_ context.Context, dst *manifest.InputBinding, val string) error {
						val = strings.TrimSpace(val)
						if val == "" {
							return errors.New("empty flag name")
						}

						dst.FlagName = val
						return nil
					},
				),
			),
		),
		func(_ context.Context, dst *manifest.InputDefinition, val *manifest.InputBinding) error {
			dst.Binding = val
			return nil
		},
	),
	Field("items",
		Pointer(listTypeSchema),
		func(_ context.Context, dst *manifest.InputDefinition, val *manifest.TypeSchema) error {
			if dst.Schema.Type == manifest.ValueTypeList {
				if val == nil {
					return errors.New("missing array element type definition")
				}
			} else if val != nil {
				return errors.New(`"items" property should be present only when "type" is "list"`)
			}

			dst.Schema.Items = val
			return nil
		},
	),
).Validation(func(ctx context.Context, n ast.Node, dst *manifest.InputDefinition) error {
	c, err := getLoaderContext(ctx)
	if err != nil {
		return err
	}

	rng, offset := GetNodeRange(n)
	dst.Location = manifest.ReferenceLocation{
		FileName: c.filePath,
		Range:    rng,
		Offset:   offset,
	}

	return nil
})

type defaultValVisitor struct {
	inputDef *manifest.InputDefinition
}

func newDefaultValVisitor(inputDef *manifest.InputDefinition) defaultValVisitor {
	return defaultValVisitor{inputDef: inputDef}
}

func (v defaultValVisitor) readOtherNode(ctx context.Context, opts *TraverseOpts, loc *manifest.ReferenceLocation, node ast.Node) (*manifest.TypedLazyValue, parsetypes.Diagnostics) {
	var itemReader ValueVisitor[any]
	switch v.inputDef.Schema.Type {
	case manifest.ValueTypeInt:
		itemReader = IntoAny(Int[int64]())
	case manifest.ValueTypeBool:
		itemReader = IntoAny(Bool())
	default:
		return nil, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(
				opts.FileName, node,
				fmt.Errorf("unsupported value type in this context: %s", v.inputDef.Schema),
			),
		}
	}

	val, diags := itemReader.VisitItem(ctx, opts, node)
	if diags.HasError() {
		return nil, diags
	}

	return &manifest.TypedLazyValue{
		Type: v.inputDef.Schema.Type,
		Value: manifest.LazyValue{
			Location: loc,
			Value: manifest.AnySpec{
				LiteralSpec: &manifest.LiteralSpec{
					Value: val,
				},
			},
		},
	}, nil
}

func (v defaultValVisitor) readString(ctx context.Context, n *ast.StringNode, loc manifest.ReferenceLocation) (*manifest.TypedLazyValue, error) {
	var parseFunc func(string) (any, error)
	switch typ := v.inputDef.Schema.Type; typ {
	case manifest.ValueTypeString:
		break
	case manifest.ValueTypeDate:
		parseFunc = func(val string) (any, error) {
			return time.Parse(val, v.inputDef.Schema.DateFormatOrDefault())
		}
	case manifest.ValueTypeDuration:
		parseFunc = func(val string) (any, error) {
			return time.ParseDuration(val)
		}
	default:
		return nil, fmt.Errorf("expected value of type %s but got string", typ)
	}

	exp, err := expr.Parse(n.Value)
	if err != nil {
		return nil, err
	}

	typedVal := &manifest.TypedLazyValue{
		Type: v.inputDef.Schema.Type,
		Value: manifest.LazyValue{
			Location: &loc,
		},
	}

	if exp.Evaluable() {
		typedVal.Value.Value = manifest.AnySpec{
			BindingSpec: &manifest.BindingSpec{
				Location: loc,
				Expr:     exp,
			},
		}
		return typedVal, nil
	}

	// should never return errors as isn't evaluable
	rawVal, err := exp.ByteString(ctx, expr.EvalContext{})
	if err != nil {
		return nil, err
	}

	var outVal any = string(rawVal)
	if parseFunc != nil {
		outVal, err = parseFunc(string(rawVal))
		if err != nil {
			return nil, err
		}
	}

	typedVal.Value.Value = manifest.AnySpec{
		LiteralSpec: &manifest.LiteralSpec{
			Value: outVal,
		},
	}
	return typedVal, nil
}

func (v defaultValVisitor) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) (*manifest.TypedLazyValue, parsetypes.Diagnostics) {
	rng, offset := GetNodeRange(node)
	loc := manifest.ReferenceLocation{
		FileName: opts.FileName,
		Range:    rng,
		Offset:   offset,
	}

	// If string - check for template expression inside
	if n, ok := IntoStringNode(node); ok {
		strVal, err := v.readString(ctx, n, loc)
		if err != nil {
			return nil, parsetypes.Diagnostics{
				newErrDiagnosticFromNode(opts.FileName, n, err),
			}
		}

		return strVal, nil
	}

	return v.readOtherNode(ctx, opts, &loc, node)
}

func intoAny[T any](_ context.Context, _ ast.Node, s T) (any, error) {
	return s, nil
}
