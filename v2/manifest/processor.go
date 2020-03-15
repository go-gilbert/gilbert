package manifest

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const (
	blockMixin  = "mixin"
	blockTask   = "task"
	blockParam  = "param"
	blockAction = "action"

	attrParamRequired = "required"
	attrParamType     = "type"
	attrParamDefault  = "default"
)

// extractTasksAndMixins builds tasks and mixins definitions out of HCL file blocks
func extractTasksAndMixins(ctx *hcl.EvalContext, blocks hclsyntax.Blocks) (Tasks, Mixins, hcl.Diagnostics) {
	tasks := make(Tasks)
	mixins := make(Mixins)

	for _, block := range blocks {
		switch block.Type {
		case blockTask, blockMixin:
			// get task/mixin block name and description
			label, desc, diags := readBlockLabels(block)
			if diags != nil {
				return nil, nil, diags
			}

			// collect block call parameters (also will return rest not-related blocks)
			params, restBlocks, diags := readBlockParams(block, ctx)
			if diags != nil {
				return nil, nil, diags
			}

			container := JobsContainer{
				Name:        label,
				Description: desc,
				Parameters:  params,
				blocks:      restBlocks,
				Context: Context{
					ctx: ctx,
				},
			}

			if block.Type == blockTask {
				tasks[label] = newTask(container)
			} else {
				mixins[label] = newMixin(container)
			}
		default:
			return nil, nil, NewDiagnosticsFromPosition(
				block.Range(), "unknown block %q", block.Type,
			)
		}
	}

	return tasks, mixins, nil
}

func readBlockParams(b *hclsyntax.Block, ctx *hcl.EvalContext) (Parameters, hclsyntax.Blocks, hcl.Diagnostics) {
	// place non-param blocks here
	otherBlocks := make(hclsyntax.Blocks, 0, len(b.Body.Blocks))
	params := make(Parameters)

	for _, block := range b.Body.Blocks {
		if block.Type != blockParam {
			// ignore non-param blocks
			otherBlocks = append(otherBlocks, block)
			continue
		}

		label, desc, diags := readBlockLabels(block)
		if diags != nil {
			return nil, nil, diags
		}

		p := Parameter{
			Name:        label,
			Description: desc,
		}

		// check if parameter is required
		p.Required, diags = getBoolAttrValue(block.Body, attrParamRequired, false)
		if diags != nil {
			return nil, nil, diags
		}

		// get parameter type
		p.Type, diags = getTypeAttrValue(block.Body, attrParamType)
		if diags != nil {
			return nil, nil, diags
		}

		// get default parameter value
		p.DefaultValue, diags = getAttrValue(block.Body, ctx, attrParamDefault, p.Type, p.Required)
		if diags != nil {
			return nil, nil, diags
		}

		params[p.Name] = p
	}

	return params, otherBlocks, nil
}

// readBlockLabels extracts block label and description
func readBlockLabels(b *hclsyntax.Block) (label, description string, diags hcl.Diagnostics) {
	labelsLen := len(b.Labels)
	if labelsLen == 0 {
		return label, description, NewDiagnosticsFromPosition(
			b.LabelRanges[0],
			"block %[1]q should have %[1]s name", b.Type,
		)
	}

	label = b.Labels[0]
	if labelsLen > 1 {
		description = b.Labels[1]
	}

	return label, description, nil
}

func appendAttrsToContext(attrs hclsyntax.Attributes, ctx *hcl.EvalContext) hcl.Diagnostics {
	// first process scalar attributes (that don't contain variable references)
	// and them any values with references
	var nonScalarAttrs []*hclsyntax.Attribute

	for attrName, attr := range attrs {
		if attrName == propImports {
			// TODO: process imports
			continue
		}

		refs := hclsyntax.Variables(attr.Expr)
		if len(refs) > 0 {
			nonScalarAttrs = append(nonScalarAttrs, attr)
			continue
		}

		attrVal, diags := attr.Expr.Value(ctx)
		if diags != nil {
			return diags
		}

		ctx.Variables[attrName] = attrVal
	}

	// Process values with references
	for _, attr := range nonScalarAttrs {
		attrVal, diags := attr.Expr.Value(ctx)
		if diags != nil {
			return diags
		}

		ctx.Variables[attr.Name] = attrVal
	}

	return nil
}
