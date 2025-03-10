package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/spf13/pflag"
)

var _ pflag.Value = (*inputFlagBinding)(nil)

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

type inputFlagContext struct {
	evalContext      expr.EvalContext
	envVars          map[string]string
	dstScope         *scope.Scope
	inputDiagnostics *inputDiagnostics
}

type inputFlagBinding struct {
	inputDef *manifest.InputDefinition
	flagCtx  inputFlagContext
	dirty    bool
	err      error
}

func newInputFlagBinding(def *manifest.InputDefinition, flagCtx inputFlagContext) *inputFlagBinding {
	return &inputFlagBinding{
		inputDef: def,
		flagCtx:  flagCtx,
	}
}

func (i *inputFlagBinding) getDoc(isGlobal bool) string {
	if len(i.inputDef.Doc) != 0 {
		return strings.Join(i.inputDef.Doc, "\n")
	}

	if isGlobal {
		return fmt.Sprintf("Set global workflow input %q", i.inputDef.Name)
	}

	return fmt.Sprintf("Set task input %q", i.inputDef.Name)
}

func (i *inputFlagBinding) flagName() string {
	binding := i.inputDef.Binding
	if i.inputDef.Binding != nil && binding.FlagName != "" {
		return binding.FlagName
	}

	// TODO: convert input name from pascal|snake case into kebab case
	return i.inputDef.Name
}

func (i *inputFlagBinding) isRequired() bool {
	if i.inputDef.DefaultValue != nil {
		return false
	}

	return i.inputDef.Binding == nil || i.inputDef.Binding.EnvVarName == ""
}

func (i *inputFlagBinding) initDefaultValue(ctx context.Context) {
	if err := i.initDefaultFromDef(ctx); err != nil {
		i.err = err
		return
	}

	if binding := i.inputDef.Binding; binding != nil && binding.EnvVarName != "" {
		if err := i.setValueFromInput(i.flagCtx.envVars[binding.EnvVarName]); err != nil {
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

	i.flagCtx.dstScope.Inputs[i.inputDef.Name] = zeroValue
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

	i.flagCtx.dstScope.Inputs[i.inputDef.Name] = parsedValue
	return nil
}

func (i *inputFlagBinding) initDefaultFromDef(ctx context.Context) error {
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

	val, err := i.inputDef.DefaultValue.Value.Expand(ctx, i.flagCtx.evalContext)
	if err != nil {
		var diags parsetypes.Diagnostics
		if errors.Is(err, &diags) {
			i.flagCtx.inputDiagnostics.add(diags...)
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
	i.flagCtx.dstScope.Inputs[key] = castedVal
	return nil
}

func (i *inputFlagBinding) addInputError(err error) error {
	if err == nil {
		return nil
	}

	i.flagCtx.inputDiagnostics.addInputError(i.inputDef, err)
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

	i.flagCtx.dstScope.Inputs[i.inputDef.Name] = val
	return nil
}

func (i *inputFlagBinding) String() string {
	val, ok := i.flagCtx.dstScope.Inputs[i.inputDef.Name]
	if !ok || val == nil {
		return ""
	}

	// TODO: make this in a proper way
	var strVal string
	switch i.inputDef.Type.Format {
	case manifest.ValueFormatInvalid:
		break
	case manifest.ValueFormatDate:
		if dt, ok := val.(time.Time); ok {
			strVal = dt.Format(i.inputDef.Type.DateFormatOrDefault())
		}
	case manifest.ValueFormatDuration:
		if dur, ok := val.(time.Duration); ok {
			strVal = dur.String()
		}
	case manifest.ValueFormatURL:
		if uri, ok := val.(*url.URL); ok {
			strVal = uri.String()
		}
	}

	if strVal == "" {
		return fmt.Sprint(val)
	}

	return strconv.Quote(strVal)
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
		return fmt.Errorf("missing required input %q", i.inputDef.Name)
	}

	return nil
}

func (i *inputFlagBinding) Type() string {
	return i.inputDef.Type.String()
}
