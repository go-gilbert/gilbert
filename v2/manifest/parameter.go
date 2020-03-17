package manifest

import (
	"github.com/zclconf/go-cty/cty"
)

const arrayTypePrefix = "[]"

var (
	emptyDefaultValue = cty.NilVal
)

type Parameters map[string]Parameter

type Parameter struct {
	Name         string
	Type         cty.Type
	Description  string
	DefaultValue cty.Value
}

func (p Parameter) IsRequired() bool {
	return p.DefaultValue.IsNull()
}
