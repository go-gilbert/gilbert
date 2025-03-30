package inputflag

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-gilbert/gilbert/internal/v2/log"
	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/internal/v2/manifest/expr"
	"github.com/go-gilbert/gilbert/internal/v2/scope"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type dirtyFlag uint8

const (
	valueClean dirtyFlag = iota
	valueDefault
	valueDirty
)

type DiagnosticsCollector struct {
	HasErrors   bool
	Diagnostics parsetypes.Diagnostics
}

func NewDiagnosticsCollector() *DiagnosticsCollector {
	return &DiagnosticsCollector{}
}

func (i *DiagnosticsCollector) Append(newDiags ...*parsetypes.Diagnostic) {
	if !i.HasErrors {
		i.HasErrors = parsetypes.HasErrorDiagnostics(newDiags)
	}

	i.Diagnostics = append(i.Diagnostics, newDiags...)
}

func (i *DiagnosticsCollector) AddInputError(def *manifest.InputDefinition, err error) {
	i.HasErrors = true
	i.Diagnostics = append(i.Diagnostics, &parsetypes.Diagnostic{
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: def.Location.FileName,
		Range:    def.Location.Range,
		Offset:   def.Location.Offset,
		Err:      err,
	})
}

func (i *DiagnosticsCollector) AddErrorAtLocation(loc *manifest.ReferenceLocation, err error) {
	i.HasErrors = true
	i.Diagnostics = append(i.Diagnostics, &parsetypes.Diagnostic{
		Severity: parsetypes.DiagnosticSeverityError,
		FileName: loc.FileName,
		Range:    loc.Range,
		Offset:   loc.Offset,
		Err:      err,
	})
}

type inputFlagContext struct {
	evalContext      expr.EvalContext
	envVars          map[string]string
	dstScope         *scope.Scope
	inputDiagnostics *DiagnosticsCollector
}

// inputBindingBase contains mutual components and boilerplate code for all input binding implementations.
type inputBindingBase struct {
	logger      *log.Logger
	inputDef    *manifest.InputDefinition
	flagCtx     inputFlagContext
	dirtyStatus dirtyFlag
	err         error
}

func newInputBindingBase(logger *log.Logger, inputDef *manifest.InputDefinition, flagCtx inputFlagContext) inputBindingBase {
	return inputBindingBase{
		logger:   logger,
		inputDef: inputDef,
		flagCtx:  flagCtx,
	}
}

func (i *inputBindingBase) Type() string {
	return i.inputDef.Schema.String()
}

func (i *inputBindingBase) getDoc(isGlobal bool) string {
	if len(i.inputDef.Doc) != 0 {
		return strings.Join(i.inputDef.Doc, "\n")
	}

	if isGlobal {
		return fmt.Sprintf("Set global workflow input %q", i.inputDef.Name)
	}

	return fmt.Sprintf("Set task input %q", i.inputDef.Name)
}

func (i *inputBindingBase) flagName() string {
	binding := i.inputDef.Binding
	if i.inputDef.Binding != nil && binding.FlagName != "" {
		return binding.FlagName
	}

	// TODO: convert input name from pascal|snake case into kebab case
	return i.inputDef.Name
}

func (i *inputBindingBase) isRequired() bool {
	binding := i.inputDef.Binding
	if binding != nil && binding.EnvVarName != "" {
		// not required if there is env default value.
		v, ok := i.flagCtx.envVars[binding.EnvVarName]
		if ok && v != "" {
			return false
		}
	}

	return i.inputDef.IsRequired()
}

func (i *inputBindingBase) addInputError(err error) error {
	if err == nil {
		return nil
	}

	i.flagCtx.inputDiagnostics.AddInputError(i.inputDef, err)
	return err
}

func (i *inputBindingBase) error() error {
	// return an error only if flag loaded from a default value.
	// on other case - error was already raised during flag value parsing in Set().
	if i.dirtyStatus == valueDirty {
		return nil
	}

	return i.err
}

func (i *inputBindingBase) initZeroValue() {
	i.flagCtx.dstScope.Inputs[i.inputDef.Name] = manifest.NewZeroValue(i.inputDef.Schema.Type)
}

func (i *inputBindingBase) initDefaultFromDef(ctx context.Context) error {
	if i.inputDef.DefaultValue == nil {
		i.initZeroValue()
		return nil
	}

	if !i.inputDef.DefaultValue.IsType(i.inputDef.Schema) {
		return i.addInputError(errors.New("default value type doesn't match input type"))
	}

	val, err := i.inputDef.DefaultValue.Value.Expand(ctx, i.flagCtx.evalContext)
	if err != nil {
		var diags parsetypes.Diagnostics
		if errors.Is(err, &diags) {
			i.flagCtx.inputDiagnostics.Append(diags...)
			return fmt.Errorf("failed to expand default value inside parameter %q", i.inputDef.Name)
		}

		return i.addInputError(fmt.Errorf("failed to expand default value inside parameter %q", i.inputDef.Name))
	}

	i.dirtyStatus = valueDefault
	var castedVal any
	switch t := i.inputDef.Schema.Type; t {
	// Value for parseable types can be a Go value (e.g. time.Duration) or a raw string.
	//
	// Scenarios:
	// 	- Value is explicitly declared in a workflow and parsed by yamllloader into a Go value.
	//	- Value is dynamic but returns a Go value.
	//  - Value is dynamic but returns a string (e.g "$(date -Ihours)")
	//
	// If value is string - parse it. Otherwise - type check.
	case manifest.ValueTypeDate:
		castedVal, err = valueToDate(val, i.inputDef.Schema.DateFormatOrDefault())
	case manifest.ValueTypeDuration:
		castedVal, err = valueToDuration(val)

	// Other types
	case manifest.ValueTypeString:
		castedVal, err = parsetypes.AnyToString(val)
	case manifest.ValueTypeFloat:
		castedVal, err = parsetypes.AnyToFloat(val)
	case manifest.ValueTypeInt:
		castedVal, err = parsetypes.AnyToInt(val)
	case manifest.ValueTypeBool:
		castedVal, err = parsetypes.AnyToBool(val)
	case manifest.ValueTypeList:
		castedVal, err = parsetypes.AnyToList(val)
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

func valueToDate(v any, dateFormat string) (any, error) {
	var strVal string
	switch t := v.(type) {
	case string:
		strVal = t
	case []byte:
		strVal = string(t)
	case time.Time:
		return t, nil
	default:
		return nil, fmt.Errorf("expected value of type date but got %T", t)
	}

	dt, err := time.Parse(dateFormat, strVal)
	if err != nil {
		err = fmt.Errorf("failed to parse %q as date: %w", strVal, err)
	}

	return dt, err
}

func valueToDuration(v any) (any, error) {
	var strVal string
	switch t := v.(type) {
	case string:
		strVal = t
	case []byte:
		strVal = string(t)
	case time.Duration:
		return t, nil
	default:
		return nil, fmt.Errorf("expected value of type date but got %T", t)
	}

	dur, err := time.ParseDuration(strVal)
	if err != nil {
		err = fmt.Errorf("failed to parse %q as duration: %w", strVal, err)
	}

	return dur, err
}
