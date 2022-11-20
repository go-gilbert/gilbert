package spec

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type StepType uint

const (
	StepTypeInvalid StepType = iota
	StepTypeCallTask
	StepTypeAction
	StepTypeMixin
)

func (stepType StepType) String() string {
	switch stepType {
	case StepTypeCallTask:
		return "task"
	case StepTypeAction:
		return "action"
	case StepTypeMixin:
		return "mixin"
	default:
		return ""
	}
}

func StepTypeFromString(str string) StepType {
	switch str {
	case "task":
		return StepTypeCallTask
	case "action":
		return StepTypeAction
	case "mixin":
		return StepTypeMixin
	default:
		return StepTypeInvalid
	}
}

type Step struct {
	BlockHeader
	Type       StepType
	Attributes hclsyntax.Attributes
	Blocks     hclsyntax.Blocks
}

func ParseStep(block *hclsyntax.Block, ctx *hcl.EvalContext) (*Step, hcl.Diagnostics) {
	stepType := StepTypeFromString(block.Type)
	if stepType == StepTypeInvalid {
		return nil, newDiagnosticError(block.TypeRange,
			"invalid task step block %q", block.Type)
	}

	header, err := ParseHeader(block, ctx, WithBlockKey(true))
	if err != nil {
		return nil, err
	}

	return &Step{
		BlockHeader: header,
		Type:        stepType,
		Attributes:  block.Body.Attributes,
		Blocks:      block.Body.Blocks,
	}, nil
}
