package cmd

import (
	"fmt"
	"github.com/go-gilbert/gilbert/v2/manifest"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

type stringInjector struct {
	name     string
	required bool
	val      *string
}

func (ij stringInjector) Required() bool {
	return ij.required
}

func (ij stringInjector) ValueEmpty() bool {
	if ij.val == nil {
		return true
	}

	return *ij.val == ""
}

func (ij stringInjector) InjectParameter(ctx *hcl.EvalContext) error {
	ctx.Variables[ij.name] = cty.StringVal(*ij.val)
	return nil
}

type floatInjector struct {
	name     string
	required bool
	val      *float64
}

func (ij floatInjector) Required() bool {
	return ij.required
}

func (ij floatInjector) InjectParameter(ctx *hcl.EvalContext) error {
	ctx.Variables[ij.name] = cty.NumberFloatVal(*ij.val)
	return nil
}

func (ij floatInjector) ValueEmpty() bool {
	if ij.val == nil {
		return true
	}

	return *ij.val == 0
}

type boolInjector struct {
	name string
	val  *bool
}

func (ij boolInjector) Required() bool {
	// bool params are always optional
	return false
}

func (ij boolInjector) ValueEmpty() bool {
	return ij.val == nil
}

func (ij boolInjector) InjectParameter(ctx *hcl.EvalContext) error {
	ctx.Variables[ij.name] = cty.BoolVal(*ij.val)
	return nil
}

type CtxParamInjector interface {
	InjectParameter(ctx *hcl.EvalContext) error
	Required() bool
	ValueEmpty() bool
}

// InjectableParams holds set of task parameters from command line
// that should be injected into task's context.
type InjectableParams map[string]CtxParamInjector

// CheckParams checks if all required task parameters are satisfied
//
// Necessary, because cobra.MarkFlagRequired() doesn't seem to work.
func (ip InjectableParams) CheckParams() error {
	for name, ij := range ip {
		// optional params always have non-nil value
		if ij.ValueEmpty() && ij.Required() {
			return fmt.Errorf("missing required task parameter %q", name)
		}
	}

	return nil
}

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
		required := param.IsRequired()

		switch param.Type {
		case cty.String:
			var defaultVal string
			if !required {
				defaultVal = param.DefaultValue.AsString()
			}

			params[name] = stringInjector{
				name:     name,
				required: required,
				val:      flags.String(name, defaultVal, param.Description),
			}
		case cty.Number:
			var defaultVal float64
			if !required {
				defaultVal, _ = param.DefaultValue.AsBigFloat().Float64()
			}

			params[name] = floatInjector{
				name:     name,
				required: required,
				val:      flags.Float64(name, defaultVal, param.Description),
			}
		case cty.Bool:
			defaultVal := false
			if !required {
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

		// Doesn't work, CheckParams() workaround used instead
		//if err := c.MarkFlagRequired(name); err != nil {
		//	return nil, err
		//}
	}

	//err := c.Flags().Parse(args)
	err := flags.Parse(args)
	if err != nil {
		// Add error description if it's unknown flag error
		if isUnknownFlagError(err) {
			return nil, fmt.Errorf(
				"%[1]s\n\nUnknown task parameter or command flag.\n"+
					`Check task parameters with "%[2]s inpect %[3]s" command or run "%[2]s help run"`,
				err, BinName, t.Name,
			)
		}

		return nil, err
	}

	if err := params.CheckParams(); err != nil {
		return nil, fmt.Errorf(
			"%s.\n\nCheck required task parameters with \"%s inspect %s\"",
			err, BinName, t.Name,
		)
	}

	return params, nil
}

func isUnknownFlagError(err error) bool {
	return strings.HasPrefix(err.Error(), "unknown flag")
}
