package spec

import (
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
		return nil, newDiagnosticError(block.Range(),
			"attribute %q is required if parameter is not optional", paramDefaultAttr)
	}

	if typeAttr, ok := attrs[paramTypeAttr]; ok {
		typ, err := ctyTypeFromExpression(typeAttr.Expr)
		if err != nil {
			return nil, err
		}

		param.Type = typ
	}

	if param.Type == cty.NilType && param.DefaultValue == nil {
		param.Type = cty.String
	}

	optSpec, err := traverseParamBlocks(block.Body.Blocks, ctx)
	if err != nil {
		return nil, err
	}

	param.Option = optSpec
	return &param, nil
}

func traverseParamBlocks(blocks hclsyntax.Blocks, ctx *hcl.EvalContext) (*FlagSpec, hcl.Diagnostics) {
	if len(blocks) == 0 {
		return nil, nil
	}

	var specBlock *hclsyntax.Block
	for _, block := range blocks {
		if block.Type != blockTypeParamFlag {
			return nil, newDiagnosticError(block.DefRange(), "unsupported block %q",
				block.Type)
		}

		if specBlock != nil {
			return nil, newDiagnosticError(block.DefRange(),
				"duplicate %q block, previous block was defined at line %d",
				block.Type, specBlock.TypeRange.Start.Line)
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
		return nil, newDiagnosticError(specBlock.DefRange(),
			"missing %q attribute", paramFlagName)
	}

	flagName, err := unmarshalAttr[string](flagNameAttr, ctx)
	if err != nil {
		return nil, err
	}

	flagName = strings.TrimSpace(flagName)
	if flagName == "" {
		return nil, newDiagnosticError(flagNameAttr.Range,
			"empty required attribute %q", flagNameAttr.Name)
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
