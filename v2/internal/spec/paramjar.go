package spec

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Variables = map[string]cty.Value

// paramsContextKey is attribute key for params in hcl.EvalContext
const paramsContextKey = "params"

// ParamJar manages execution parameters in hcl.EvalContext
type ParamJar struct {
	values Variables
}

// NewParamJar returns a new params jar.
func NewParamJar(size int) *ParamJar {
	return &ParamJar{values: make(Variables, size)}
}

// Set sets a value in a jar.
func (j *ParamJar) Set(param string, val cty.Value) {
	j.values[param] = val
}

// Get returns a value from a jar
func (j *ParamJar) Get(param string) (cty.Value, bool) {
	v, ok := j.values[param]
	if !ok {
		return cty.NilVal, false
	}

	return v, ok
}

// MapContext maps params to passed hcl.EvalContext
func (j *ParamJar) MapContext(ctx *hcl.EvalContext) {
	if len(j.values) == 0 {
		return
	}

	if ctx.Variables == nil {
		ctx.Variables = make(Variables)
	}

	ctx.Variables[paramsContextKey] = cty.ObjectVal(j.values)
}
