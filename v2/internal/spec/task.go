package spec

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const (
	paramTaskAsync  = "async"
	paramTaskMatrix = "matrix"
)

type Tasks map[string]*Task

type Task struct {
	BlockHeader
	TaskAttributes
	Params Params
	Vars   *Vars
	Steps  []*Step
}

type TaskAttributes struct {
	Async  hcl.Expression
	Matrix hcl.Expression
}

func ParseTask(block *hclsyntax.Block, ctx *hcl.EvalContext) (*Task, hcl.Diagnostics) {
	header, err := ParseHeader(block, ctx)
	if err != nil {
		return nil, err
	}

	attrs, err := traverseTaskAttributes(block.Body.Attributes)
	if err != nil {
		return nil, err
	}

	task := &Task{
		BlockHeader:    header,
		TaskAttributes: attrs,
		Params:         make(Params),
		Steps:          make([]*Step, 0, len(block.Body.Blocks)),
	}

	if err := traverseTaskBlocks(task, block.Body.Blocks, ctx); err != nil {
		return nil, err
	}

	return task, nil
}

func traverseTaskBlocks(task *Task, blocks hclsyntax.Blocks, ctx *hcl.EvalContext) hcl.Diagnostics {
	for _, block := range blocks {
		switch block.Type {
		case varsBlockName:
			vars, err := ParseVars(block.AsHCLBlock())
			if err != nil {
				return err
			}

			task.Vars = vars
		case paramBlockName:
			param, err := ParseParam(block, ctx)
			if err != nil {
				return err
			}

			task.Params[param.Name] = param
		default:
			step, err := ParseStep(block, ctx)
			if err != nil {
				return err
			}

			task.Steps = append(task.Steps, step)
		}
	}

	return nil
}

func traverseTaskAttributes(attrs hclsyntax.Attributes) (TaskAttributes, hcl.Diagnostics) {
	var out TaskAttributes
	if len(attrs) == 0 {
		return out, nil
	}

	for attrName, attr := range attrs {
		switch attrName {
		case paramTaskAsync:
			out.Async = attr.Expr
		case paramTaskMatrix:
			out.Matrix = attr.Expr
		default:
			return out, newDiagnosticError(attr.Range(),
				"unsupported attribute %q in task block", attrName)
		}
	}

	return out, nil
}
