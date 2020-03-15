package manifest

import (
	"github.com/zclconf/go-cty/cty"
)

const arrayTypePrefix = "[]"

type Parameters map[string]Parameter

type Parameter struct {
	Name         string
	Type         cty.Type
	Description  string
	Required     bool
	DefaultValue cty.Value
}
