package spec

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"runtime"
)

type RunnerSpec struct {
	Version   string
	Build     string
	CommitSHA string
}

func (s *RunnerSpec) CtyValue() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"version": cty.StringVal(s.Version),
		"build":   cty.StringVal(s.Build),
		"commit":  cty.StringVal(s.CommitSHA),
	})
}

type PlatformSpec struct {
	OS   string
	Arch string
}

func (s PlatformSpec) CtyValue() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"os":   cty.StringVal(s.OS),
		"arch": cty.StringVal(s.Arch),
	})
}

type RootContext struct {
	Runner   RunnerSpec
	Platform PlatformSpec
}

func NewRootContext() *RootContext {
	return &RootContext{
		Platform: PlatformSpec{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}
}

func (ctx *RootContext) Map(evalCtx *hcl.EvalContext) {
	evalCtx.Functions = builtinFunctions
	evalCtx.Variables = map[string]cty.Value{
		"platform": ctx.Platform.CtyValue(),
		"gilbert":  ctx.Runner.CtyValue(),
	}
}
