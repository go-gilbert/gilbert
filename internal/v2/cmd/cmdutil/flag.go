package cmdutil

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type inputDiagnostics struct {
	hasErrors bool
	diags     parsetypes.Diagnostics
}

func (i *inputDiagnostics) add(newDiags ...*parsetypes.Diagnostic) {
	if !i.hasErrors {
		i.hasErrors = parsetypes.HasErrorDiagnostics(newDiags)
	}

	i.diags = append(i.diags, newDiags...)
}

func (i *inputDiagnostics) addInputError(def *manifest.InputDefinition, err error) {
	i.hasErrors = true
	i.diags = append(i.diags, &parsetypes.Diagnostic{
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: def.Location.FileName,
		Range:    def.Location.Range,
		Offset:   def.Location.Offset,
		Err:      err,
	})
}

type inputFlagBinding struct {
	inputDef    *manifest.InputDefinition
	evalContext expr.EvalContext
	envVars     map[string]string
	writeScope  *scope.Scope
	diags       *inputDiagnostics
	dirty       bool
	err         error
}

func (i *inputFlagBinding) initDefaultValue() {
	if err := i.initDefaultFromDef(); err != nil {
		i.err = err
		return
	}

	if binding := i.inputDef.Binding; binding != nil && binding.EnvVarName != "" {
		if err := i.setValueFromInput(i.envVars[binding.EnvVarName]); err != nil {
			i.err = err
			return
		}
	}
}

func (i *inputFlagBinding) initZeroValue() {
	var zeroValue any
	switch t := i.inputDef.Type.Type; t {
	case manifest.ValueTypeString:
		switch i.inputDef.Type.Format {
		case manifest.ValueFormatDate:
			zeroValue = time.Now()
		case manifest.ValueFormatDuration:
			zeroValue = time.Duration(0)
		case manifest.ValueFormatURL:
			zeroValue = &url.URL{}
		default:
			zeroValue = ""
		}
	case manifest.ValueTypeFloat:
		zeroValue = float64(0)
	case manifest.ValueTypeInt:
		zeroValue = int64(0)
	case manifest.ValueTypeBool:
		zeroValue = false
	default:
		zeroValue = nil
	}

	i.writeScope.Inputs[i.inputDef.Name] = zeroValue
}

func (i *inputFlagBinding) setValueFromInput(val string) error {
	if val == "" {
		return nil
	}

	i.dirty = true
	var (
		parsedValue any
		err         error
	)
	switch t := i.inputDef.Type.Type; t {
	case manifest.ValueTypeInt:
		parsedValue, err = strconv.ParseInt(val, 10, 64)
	case manifest.ValueTypeFloat:
		parsedValue, err = strconv.ParseFloat(val, 64)
	case manifest.ValueTypeBool:
		parsedValue, err = strconv.ParseBool(val)
	case manifest.ValueTypeString:
		parsedValue, err = i.inputDef.Type.ParseString(val)
		if err != nil {
			return fmt.Errorf("cannot parse %q as %s: %w", val, i.inputDef.Type.Format, err)
		}

	default:
		// TODO: support list of scalars
		return fmt.Errorf("cannot parse string %q as %s", val, t)
	}

	if err != nil {
		return err
	}

	i.writeScope.Inputs[i.inputDef.Name] = parsedValue
	return nil
}

func (i *inputFlagBinding) initDefaultFromDef() error {
	if i.inputDef.DefaultValue == nil {
		i.initZeroValue()
		return nil
	}

	if !i.inputDef.DefaultValue.IsType(i.inputDef.Type) {
		return i.addInputError(errors.New("default value type doesn't match input type"))
	}

	// TODO: support list of scalar types.
	if i.inputDef.Type.Type == manifest.ValueTypeList {
		return i.addInputError(errors.New("input of type list cannot be bound to flags"))
	}

	val, err := i.inputDef.DefaultValue.Value.Expand(i.evalContext)
	if err != nil {
		var diags parsetypes.Diagnostics
		if errors.Is(err, &diags) {
			i.diags.add(diags...)
			return fmt.Errorf("failed to expand default value inside parameter %q", i.inputDef.Name)
		}

		return i.addInputError(fmt.Errorf("failed to expand default value inside parameter %q", i.inputDef.Name))
	}

	i.dirty = true
	var castedVal any
	switch t := i.inputDef.Type.Type; t {
	case manifest.ValueTypeString:
		err = i.initStringValue(val)
		return i.addInputError(err)
	case manifest.ValueTypeFloat:
		castedVal, err = parsetypes.AnyToFloat(val)
	case manifest.ValueTypeInt:
		castedVal, err = parsetypes.AnyToInt(val)
	case manifest.ValueTypeBool:
		castedVal, err = parsetypes.AnyToBool(val)
	default:
		return i.addInputError(fmt.Errorf("unsupported default value type: %s", t))
	}

	if err != nil {
		return i.addInputError(err)
	}

	key := i.inputDef.Name
	i.writeScope.Inputs[key] = castedVal
	return nil
}

func (i *inputFlagBinding) addInputError(err error) error {
	if err == nil {
		return nil
	}

	i.diags.addInputError(i.inputDef, err)
	return err
}

func (i *inputFlagBinding) initStringValue(rawVal any) error {
	strVal, err := parsetypes.AnyToString(rawVal)
	if err != nil {
		return err
	}

	val, err := i.inputDef.Type.ParseString(strVal)
	if err != nil {
		return fmt.Errorf("cannot parse %q as %s: %w", strVal, i.inputDef.Type.Format, err)
	}

	i.writeScope.Inputs[i.inputDef.Name] = val
	return nil
}

func (i *inputFlagBinding) String() string {
	val, ok := i.writeScope.Inputs[i.inputDef.Name]
	if !ok {
		return ""
	}

	return fmt.Sprint(val)
}

func (i *inputFlagBinding) Set(s string) error {
	if err := i.setValueFromInput(s); err != nil {
		i.err = err
		return err
	}

	i.dirty = true
	return nil
}

func (i *inputFlagBinding) validate() error {
	if i.err != nil {
		return i.err
	}

	if !i.dirty {
		return fmt.Errorf("missing ")
	}
}

func (i *inputFlagBinding) Type() string {
	return i.inputDef.Type.String()
}
