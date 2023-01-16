package cmd

import (
	"fmt"
	"github.com/go-gilbert/gilbert/v2/internal/log"
	"github.com/go-gilbert/gilbert/v2/internal/spec"
	"github.com/go-gilbert/gilbert/v2/internal/util/ctyutil"
	"github.com/go-gilbert/gilbert/v2/internal/util/hclx"
	"github.com/hashicorp/hcl/v2"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"
)

type CtyValueFlag struct {
	name    string
	jar     *spec.ParamJar
	valType cty.Type
}

// NewCtyValueFlag constructs new flag parser for cty.Value.
func NewCtyValueFlag(name string, jar *spec.ParamJar, valType cty.Type) *CtyValueFlag {
	return &CtyValueFlag{name: name, jar: jar, valType: valType}
}

func (c *CtyValueFlag) String() string {
	// TODO: check if valType.IsPrimitiveType
	val, ok := c.jar.Get(c.name)
	if !ok {
		return ""
	}
	return ctyutil.ValueToString(val)
}

func (c *CtyValueFlag) Set(s string) error {
	val, err := ctyutil.ValueFromString(c.valType, s)
	if err != nil {
		return err
	}

	c.jar.Set(c.name, val)
	return nil
}

func (c *CtyValueFlag) Type() string {
	return ctyutil.TypeToString(c.valType)
}

// ApplySpecToCommand appends flags and sub-commands from spec.Spec to command.
func ApplySpecToCommand(cmd *cobra.Command, s *spec.Spec, ctx *hcl.EvalContext) (*spec.ParamJar, error) {
	jar := spec.NewParamJar(len(s.Params))
	for key, param := range s.Params {
		log.Global().Debugf("mounting root param %q", key)
		if !param.Type.IsPrimitiveType() {
			return nil, hclx.NewDiagnosticError(param.Range, "Parameter type should be primitive")
		}

		if param.DefaultValue != nil {
			log.Global().Debugf("parsing root %q param default value", key)
			v, err := param.DefaultValue.Value(ctx)
			if err != nil {
				return nil, hclx.WrapDiagnostic(err, "Failed to determine default value for variable %q", key)
			}

			jar.Set(key, v)
		}

		flagName := paramFlagName(param)
		log.Global().Debugf("mounting root param %q as flag %q", key, flagName)
		cmd.PersistentFlags().Var(
			NewCtyValueFlag(key, jar, param.Type), flagName, paramDescription(param),
		)

		if param.Required {
			log.Global().Debugf("marking flag %q as required", flagName)
			if err := cmd.MarkPersistentFlagRequired(flagName); err != nil {
				return nil, fmt.Errorf("failed to set flag required for param %q: %w", param.Name, err)
			}
		}
	}

	return jar, nil
}

func paramFlagName(param *spec.Param) string {
	if param.Option == nil && param.Option.Name == "" {
		return strcase.ToKebab(param.Name)
	}

	if strcase.ToKebab(param.Option.Name) != param.Option.Name {
		log.Global().Warnf("flag name for parameter %q is not in kebab-case", param.Name)
	}

	return strcase.ToKebab(param.Name)
}

func paramDescription(param *spec.Param) string {
	if param.Description != "" {
		return param.Description
	}

	return fmt.Sprintf("sets value for parameter %q", param.Name)
}
