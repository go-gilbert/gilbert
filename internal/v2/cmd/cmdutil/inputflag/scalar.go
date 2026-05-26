package inputflag

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

var _ inputFlagBinding = (*scalarInputFlagBinding)(nil)

// scalarInputFlagBinding binds scalar inputs (numbers, strings, booleans) to command-line flags.
type scalarInputFlagBinding struct {
	inputBindingBase
}

func newScalarInputFlagBinding(logger *log.Logger, def *manifest.InputDefinition, flagCtx inputFlagContext) *scalarInputFlagBinding {
	return &scalarInputFlagBinding{
		inputBindingBase: newInputBindingBase(logger, def, flagCtx),
	}
}

func (i *scalarInputFlagBinding) IsBoolFlag() bool {
	return i.inputDef.Schema.Type == manifest.ValueTypeBool
}

func (i *scalarInputFlagBinding) checkType() error {
	typeDef := i.inputDef.Schema
	if !typeDef.Type.IsComplex() {
		return nil
	}

	if typeDef.Type == manifest.ValueTypeList {
		if typeDef.Items == nil {
			// TODO: handle in yamlloader
			return fmt.Errorf(
				"empty list item type definition for input %q (declared at %s:%s)",
				i.inputDef.Name, i.inputDef.Location.FileName, i.inputDef.Location.Range,
			)
		}

		if !typeDef.Items.Type.IsComplex() {
			// scalar lists are ok
			return nil
		}
	}

	return fmt.Errorf(
		"input of type %s cannot be mounted as a command flag (declared at %s:%s)",
		i.inputDef.Schema, i.inputDef.Location.FileName, i.inputDef.Location.Range,
	)
}

func (i *scalarInputFlagBinding) initDefaultValue(ctx context.Context) {
	if err := i.checkType(); err != nil {
		i.err = err
		return
	}

	if err := i.initDefaultFromDef(ctx); err != nil {
		i.err = err
		return
	}

	if err := i.initFromEnvVar(); err != nil {
		i.flagCtx.inputDiagnostics.AddErrorAtLocation(i.inputDef.Binding.Location.EnvVarName, err)
		i.err = err
	}
}

func (i *scalarInputFlagBinding) initFromEnvVar() error {
	binding := i.inputDef.Binding
	if binding == nil || binding.EnvVarName == "" {
		return nil
	}

	envVal, ok := i.flagCtx.envVars[binding.EnvVarName]
	if !ok {
		return nil
	}

	i.logger.Debugw(
		"reading default value for input from environment",
		log.NewField("env", binding.EnvVarName),
		log.NewField("value", envVal),
		log.NewField("input", i.inputDef.Name),
	)

	err := i.setValueFromInput(envVal, true)
	if err != nil {
		return parsetypes.NewAnnotatedError(
			fmt.Errorf(
				"cannot set default value from environment variable %q: %w",
				binding.EnvVarName, err,
			),
			"expands as %q", envVal,
		)
	}

	return nil
}

func (i *scalarInputFlagBinding) setValueFromInput(val string, isDefault bool) error {
	if val == "" {
		return nil
	}

	if isDefault {
		i.dirtyStatus = valueDefault
	} else {
		i.dirtyStatus = valueDirty
	}

	parsedValue, err := i.inputDef.Schema.ParseValue(val)
	if err != nil {
		return err
	}

	i.flagCtx.dstScope.Inputs[i.inputDef.Name] = parsedValue
	return nil
}

func (i *scalarInputFlagBinding) String() string {
	typ := i.inputDef.Schema.Type
	val, ok := i.flagCtx.dstScope.Inputs[i.inputDef.Name]
	if !ok || val == nil {
		return ""
	}

	// TODO: make this in a proper way
	var strVal string
	switch typ {
	case manifest.ValueTypeDate:
		if dt, ok := val.(time.Time); ok {
			strVal = dt.Format(i.inputDef.Schema.DateFormatOrDefault())
		}
	case manifest.ValueTypeDuration:
		if dur, ok := val.(time.Duration); ok {
			strVal = dur.String()
		}
	}

	if strVal == "" {
		return fmt.Sprint(val)
	}

	return strconv.Quote(strVal)
}

func (i *scalarInputFlagBinding) Set(s string) error {
	if err := i.setValueFromInput(s, false); err != nil {
		i.err = err
		return err
	}

	i.dirtyStatus = valueDirty
	return nil
}
