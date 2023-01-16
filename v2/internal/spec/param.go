package spec

import (
	"github.com/go-gilbert/gilbert/v2/internal/util/hclx"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

const (
	paramValidateAttr = "validate"
	paramDefaultAttr  = "default"
	paramRequiredAttr = "required"
	paramTypeAttr     = "type"

	blockTypeParamFlag = "flag"
	paramFlagName      = "name"
	paramFlagNameShort = "short"
)

type FlagSpec struct {
	Name  string
	Short string
}

type Params map[string]*Param

type Param struct {
	BlockHeader
	Type           cty.Type
	Required       bool
	Option         *FlagSpec
	DefaultValue   hcl.Expression
	ValidationExpr hcl.Expression
	Range          hcl.Range
}

func ParseParam(block *hclsyntax.Block, ctx *hcl.EvalContext) (*Param, hcl.Diagnostics) {
	header, err := ParseHeader(block, ctx)
	if err != nil {
		return nil, err
	}

	param := Param{
		BlockHeader:  header,
		Required:     false,
		DefaultValue: nil,
		Type:         cty.NilType,
		Range:        block.Range(),
	}

	attrs := block.Body.Attributes
	if defaultValAttr, ok := attrs[paramDefaultAttr]; ok {
		param.DefaultValue = defaultValAttr.Expr
		param.Type = figureOutExpressionType(defaultValAttr.Expr)
	}

	if validateAttr, ok := attrs[paramValidateAttr]; ok {
		param.ValidationExpr = validateAttr.Expr
	}

	isRequired, _, err := extractAttr[bool](paramRequiredAttr, attrs, ctx)
	param.Required = isRequired
	if err != nil {
		return nil, err
	}

	if !param.Required && param.DefaultValue == nil {
		return nil, hcl.Diagnostics{
			hclx.NewDiagnostic(block.Range(),
				hclx.WithSummary("Missing attribute type"),
				hclx.WithDetail("Attribute %q is required if parameter is not optional", paramDefaultAttr),
			),
		}
	}

	if typeAttr, ok := attrs[paramTypeAttr]; ok {
		typ, err := ctyTypeFromExpression(typeAttr.Expr)
		if err != nil {
			return nil, err
		}

		// Check if default value type equals to specified
		if param.Type != cty.NilType && typ != param.Type {
			return nil, hcl.Diagnostics{
				hclx.NewDiagnostic(typeAttr.Range(),
					hclx.WithSummary("Type mismatch"),
					hclx.WithContext(block.Range()),
					hclx.WithDetail(
						"Type specified in attribute %q is not equal to type of value in attribute %q",
						paramTypeAttr, paramDefaultAttr,
					),
				),
			}
		}

		param.Type = typ
	}

	if param.Type == cty.NilType {
		if param.DefaultValue != nil {
			// Show error if it's impossible to determine value type from default value.
			return nil, hcl.Diagnostics{
				hclx.NewDiagnostic(block.Range(),
					hclx.WithSummary("Unknown param block type"),
					hclx.WithDetail("Cannot determine parameter value type from default value. "+
						"Please specify parameter type in %q attribute", paramTypeAttr),
				),
			}
		}

		// Set default type value
		param.Type = cty.String
	}

	optSpec, err := traverseParamBlocks(block.Body.Blocks, ctx)
	if err != nil {
		return nil, err
	}

	param.Option = optSpec
	return &param, nil
}

// figureOutExpressionType returns expression value type if it's a literal expression.
//
// Returns cty.NilType on failure.
func figureOutExpressionType(expr hclsyntax.Expression) cty.Type {
	if litExp, ok := expr.(*hclsyntax.LiteralValueExpr); ok {
		return litExp.Val.Type()
	}

	return cty.NilType
}

func traverseParamBlocks(blocks hclsyntax.Blocks, ctx *hcl.EvalContext) (*FlagSpec, hcl.Diagnostics) {
	if len(blocks) == 0 {
		return nil, nil
	}

	var specBlock *hclsyntax.Block
	for _, block := range blocks {
		if block.Type != blockTypeParamFlag {
			return nil, hcl.Diagnostics{
				hclx.NewDiagnostic(block.LabelRanges[0],
					hclx.WithSummary("Unsupported block"),
					hclx.WithDetail("Unsupported block %q", block.Type),
					hclx.WithContext(block.DefRange())),
			}
		}

		if specBlock != nil {
			return nil, hcl.Diagnostics{
				hclx.NewDiagnostic(block.LabelRanges[0],
					hclx.WithSummary("Duplicate block"),
					hclx.WithDetail("Duplicate block %q, previous block was defined at %d:%d",
						block.Type, specBlock.TypeRange.Start.Line, specBlock.TypeRange.Start.Column),
					hclx.WithContext(block.DefRange())),
			}
		}

		specBlock = block
		break
	}

	attrs, err := specBlock.Body.JustAttributes()
	if err != nil {
		return nil, err
	}

	flagNameAttr, ok := attrs[paramFlagName]
	if !ok {
		return nil, hcl.Diagnostics{
			hclx.NewDiagnostic(specBlock.DefRange(),
				hclx.WithSummary("Missing attribute %q", paramFlagName)),
		}
	}

	flagName, err := unmarshalAttr[string](flagNameAttr, ctx)
	if err != nil {
		return nil, err
	}

	flagName = strings.TrimSpace(flagName)
	if flagName == "" {
		return nil, hcl.Diagnostics{
			hclx.NewDiagnostic(flagNameAttr.Range,
				hclx.WithSummary("Empty required attribute %q", flagNameAttr.Name)),
		}
	}

	flagSpec := FlagSpec{
		Name: flagName,
	}

	flagShortAttr, ok := attrs[paramFlagNameShort]
	if !ok {
		return &flagSpec, nil
	}

	flagSpec.Short, err = unmarshalAttr[string](flagShortAttr, ctx)
	return &flagSpec, err
}
