package cmd

import (
	"fmt"
	"github.com/go-gilbert/gilbert/v2/manifest"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
)

type stringInjector struct {
	name string
	val  *string
}

func (ij stringInjector) InjectParameter(ctx *hcl.EvalContext) error {
	ctx.Variables[ij.name] = cty.StringVal(*ij.val)
	return nil
}

type floatInjector struct {
	name string
	val  *float64
}

func (ij floatInjector) InjectParameter(ctx *hcl.EvalContext) error {
	ctx.Variables[ij.name] = cty.NumberFloatVal(*ij.val)
	return nil
}

type boolInjector struct {
	name string
	val  *bool
}

func (ij boolInjector) InjectParameter(ctx *hcl.EvalContext) error {
	ctx.Variables[ij.name] = cty.BoolVal(*ij.val)
	return nil
}

type CtxParamInjector interface {
	InjectParameter(ctx *hcl.EvalContext) error
}

type InjectableParams map[string]CtxParamInjector

func (ip InjectableParams) InjectParameters(ctx *hcl.EvalContext) error {
	if len(ip) == 0 {
		return nil
	}
	for name, inj := range ip {
		if err := inj.InjectParameter(ctx); err != nil {
			return fmt.Errorf("failed to inject parameter %q: %w", name, err)
		}
	}
	return nil
}

func ProcessTaskFlags(t *manifest.Task, c *cobra.Command, args []string) (InjectableParams, error) {
	if len(t.Parameters) == 0 {
		return nil, nil
	}

	params := make(InjectableParams, len(t.Parameters))
	flags := c.Flags()
	for name, param := range t.Parameters {
		defValType := param.DefaultValue.Type()
		required := defValType != cty.NilType

		switch param.Type {
		case cty.String:
			var defaultVal string
			if required {
				defaultVal = param.DefaultValue.AsString()
			}

			params[name] = stringInjector{
				name: name,
				val:  flags.String(name, defaultVal, param.Description),
			}
		case cty.Number:
			var defaultVal float64
			if required {
				defaultVal, _ = param.DefaultValue.AsBigFloat().Float64()
			}

			params[name] = floatInjector{
				name: name,
				val:  flags.Float64(name, defaultVal, param.Description),
			}
		case cty.Bool:
			defaultVal := false
			if required {
				defaultVal = param.DefaultValue.True()
			}

			params[name] = boolInjector{
				name: name,
				val:  flags.Bool(name, defaultVal, param.Description),
			}
		default:
			return nil, fmt.Errorf(
				"task parameter %q of type %q cannot be passed as command flag",
				name, param.Type.FriendlyName(),
			)
		}
	}

	return params, flags.Parse(args)
}
