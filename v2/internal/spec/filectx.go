package spec

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type ProjectSpec struct {
	FileName         string
	WorkingDirectory string
}

func (s *ProjectSpec) CtyValue() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"work_dir": cty.StringVal(s.WorkingDirectory),
		"filename": cty.StringVal(s.FileName),
	})
}

type FileContext struct {
	Project ProjectSpec
}

func (ctx *FileContext) Map(evalCtx *hcl.EvalContext) {
	evalCtx.Variables = map[string]cty.Value{
		"project": ctx.Project.CtyValue(),
	}
}
