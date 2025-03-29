package inputflag

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

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

// inputBindingBase contains mutual components and boilerplate code for all input binding implementations.
type inputBindingBase struct {
	inputDef    *manifest.InputDefinition
	flagCtx     inputFlagContext
	dirtyStatus dirtyFlag
	err         error
}

func newInputBindingBase(inputDef *manifest.InputDefinition, flagCtx inputFlagContext) inputBindingBase {
	return inputBindingBase{inputDef: inputDef, flagCtx: flagCtx}
}

func (i *inputBindingBase) Type() string {
	return i.inputDef.Type.String()
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
	if i.inputDef.DefaultValue != nil {
		return false
	}

	return i.inputDef.Binding == nil || i.inputDef.Binding.EnvVarName == ""
}

func (i *inputBindingBase) addInputError(err error) error {
	if err == nil {
		return nil
	}

	i.flagCtx.inputDiagnostics.addInputError(i.inputDef, err)
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

func (i *inputBindingBase) initDefaultFromDef(ctx context.Context) error {
	if i.inputDef.DefaultValue == nil {
		i.initZeroValue()
		return nil
	}

	if !i.inputDef.DefaultValue.IsType(i.inputDef.Type) {
		return i.addInputError(errors.New("default value type doesn't match input type"))
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

	i.dirtyStatus = valueDefault
	var castedVal any
	switch t := i.inputDef.Type.Type; t {
	case manifest.ValueTypeString:
		castedVal, err = formatValueWithSchema(i.inputDef.Type, val)
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

// formatValueWithSchema consumes a string value and formats using type schema.
//
// If passed value was already formatted - return original value.
func formatValueWithSchema(typeDef manifest.TypeSchema, rawVal any) (any, error) {
	if typeDef.Format == manifest.ValueFormatInvalid {
		// If value was already expanded before
		return parsetypes.AnyToString(rawVal)
	}

	// parse value from formatted string
	switch t := rawVal.(type) {
	case string:
		return typeDef.ParseString(t)
	case []byte:
		return typeDef.ParseString(string(t))
	}

	// otherwise - value is already parsed or came from an expression. just do type check.
	switch typeDef.Format {
	case manifest.ValueFormatDate:
		if _, ok := rawVal.(time.Time); !ok {
			return nil, fmt.Errorf("value of type %T is not a %s", rawVal, typeDef.Format)
		}

	case manifest.ValueFormatDuration:
		if _, ok := rawVal.(time.Duration); !ok {
			return nil, fmt.Errorf("value of type %T is not a %s", rawVal, typeDef.Format)
		}
	default:
		return nil, fmt.Errorf("unknown value format %s", typeDef.Format)
	}

	return rawVal, nil
}
