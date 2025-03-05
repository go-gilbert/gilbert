package yamlloader

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-gilbert/gilbert/internal/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/manifest/manifest2"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	. "github.com/go-gilbert/gilbert/pkg/yamltree"
	"github.com/goccy/go-yaml/ast"
)

var listTypeSchema = Struct(
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

var inputsSchema = Map(Pointer(inputDefinitionSchema)).
	CollectDoc(func(_ context.Context, fi FieldInfo, def *manifest2.InputDefinition) *manifest2.InputDefinition {
		def.Name = fi.Key
		def.Doc = fi.Doc
		return def
	})

var inputDefinitionSchema = Struct(
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
		func(_ context.Context, _ ast.Node, dst *manifest2.InputDefinition) (ValueVisitor[*manifest2.TypedLazyValue], error) {
			return newDefaultValVisitor(dst), nil
		},
		func(ctx context.Context, dst *manifest2.InputDefinition, val *manifest2.TypedLazyValue) error {
			dst.DefaultValue = val
			return nil
		},
	),
	Field("binding",
		Pointer(
			Struct(
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
		Pointer(listTypeSchema),
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

type defaultValVisitor struct {
	inputDef *manifest2.InputDefinition
}

func newDefaultValVisitor(inputDef *manifest2.InputDefinition) defaultValVisitor {
	return defaultValVisitor{inputDef: inputDef}
}

func (v defaultValVisitor) readOtherNode(ctx context.Context, opts *TraverseOpts, loc *manifest2.ReferenceLocation, node ast.Node) (*manifest2.TypedLazyValue, parsetypes.Diagnostics) {
	var itemReader ValueVisitor[any]
	switch v.inputDef.Type {
	case manifest2.ValueTypeInt:
		itemReader = IntoAny(Int[int64]())
	case manifest2.ValueTypeBool:
		itemReader = IntoAny(Bool())
	default:
		return nil, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(
				opts.FileName, node,
				fmt.Errorf("unsupported value type in this context: %s", v.inputDef.Type),
			),
		}
	}

	val, diags := itemReader.VisitItem(ctx, opts, node)
	if diags.HasError() {
		return nil, diags
	}

	return &manifest2.TypedLazyValue{
		Type:   v.inputDef.Type,
		Format: v.inputDef.Format,
		Value: manifest2.LazyValue{
			Location: loc,
			Value: manifest2.AnySpec{
				LiteralSpec: &manifest2.LiteralSpec{
					Value: val,
				},
			},
		},
	}, nil
}

func (v defaultValVisitor) readString(n *ast.StringNode, loc manifest2.ReferenceLocation) (*manifest2.TypedLazyValue, error) {
	exp, err := expr.Parse(n.Value)
	if err != nil {
		return nil, err
	}

	typedVal := &manifest2.TypedLazyValue{
		Type:   v.inputDef.Type,
		Format: v.inputDef.Format,
		Value: manifest2.LazyValue{
			Location: &loc,
		},
	}

	if exp.Evaluable() {
		typedVal.Value.Value = manifest2.AnySpec{
			BindingSpec: &manifest2.BindingSpec{
				Location: loc,
				Expr:     exp,
			},
		}
		return typedVal, nil
	}

	if ftyp := v.inputDef.Type; ftyp != manifest2.ValueTypeString {
		return nil, fmt.Errorf("expected value of type %s but got string", ftyp.String())
	}

	// this should never happen
	rawVal, err := exp.String(expr.EvalContext{})
	if err != nil {
		return nil, err
	}

	var outVal any = string(rawVal)
	switch v.inputDef.Format {
	case manifest2.ValueFormatDate:
		outVal, err = time.Parse(string(rawVal), v.inputDef.DateFormatOrDefault())
	case manifest2.ValueFormatDuration:
		outVal, err = time.ParseDuration(string(rawVal))
	default:
		break
	}

	if err != nil {
		return nil, err
	}

	typedVal.Value.Value = manifest2.AnySpec{
		LiteralSpec: &manifest2.LiteralSpec{
			Value: outVal,
		},
	}
	return typedVal, nil
}

func (v defaultValVisitor) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) (*manifest2.TypedLazyValue, parsetypes.Diagnostics) {
	rng, offset := GetNodeRange(node)
	loc := manifest2.ReferenceLocation{
		FileName: opts.FileName,
		Range:    rng,
		Offset:   offset,
	}

	// If string - check for template expression inside
	if n, ok := IntoStringNode(node); ok {
		strVal, err := v.readString(n, loc)
		if err != nil {
			return nil, parsetypes.Diagnostics{
				newErrDiagnosticFromNode(opts.FileName, n, err),
			}
		}

		return strVal, nil
	}

	return v.readOtherNode(ctx, opts, &loc, node)
}

//func getDefaultValueVisitor(_ context.Context, node ast.Node, def *manifest2.InputDefinition) (ValueVisitor[*manifest2.LazyValue], error) {
//	if !IsPrimitiveNode(node) {
//		return nil, fmt.Errorf("expected %s but got %s", def.Type, node.Type())
//	}
//
//	if def.Type.IsComplex() {
//		//	// TODO: support lists?
//		return nil, fmt.Errorf("default values for input of type %q are not supported", def.Type)
//	}
//
//	if n, ok := IntoStringNode(node); ok {
//		exp, err := expr.Parse(n)
//		return
//	}
//
//	switch def.Type {
//	case manifest2.ValueTypeString:
//		return Transform[string, any](String(), func(ctx context.Context, node ast.Node, s string) (any, error) {
//			switch def.Format {
//			case manifest2.ValueFormatDate:
//				format := def.DateFormat
//				if format == "" {
//					format = manifest2.DefaultDateFormat
//				}
//
//				return time.Parse(format, s)
//			case manifest2.ValueFormatDuration:
//				return time.ParseDuration(s)
//			case manifest2.ValueFormatURL:
//				_, err := url.Parse(s)
//				return s, err
//			}
//			return s, nil
//		}), nil
//	case manifest2.ValueTypeInt:
//		return Transform[int64, any](Int[int64](), intoAny), nil
//	case manifest2.ValueTypeFloat:
//		return Transform[float64, any](Float[float64](), intoAny), nil
//	case manifest2.ValueTypeBool:
//		return Transform[bool, any](Bool(), intoAny), nil
//	default:
//		// TODO: support lists?
//		return nil, fmt.Errorf("default values for input of type %q are not supported", def.Type)
//	}
//}

func intoAny[T any](_ context.Context, _ ast.Node, s T) (any, error) {
	return s, nil
}
