package manifest

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-gilbert/gilbert/internal/scope"
	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

// AppendLazyEnvVarsToScope materializes and appends lazy-evaluated environment variables to a given scope.
func AppendLazyEnvVarsToScope(ctx context.Context, dst *scope.Scope, src EnvVars) parsetypes.Diagnostics {
	if len(src) == 0 {
		return nil
	}

	ep := expr.EvalParams{
		CommandProcessor: scope.NewCommandRunner(dst),
		Env:              dst,
	}

	if dst.Environment == nil {
		dst.Environment = make(map[string]string, len(src))
	}

	var diags parsetypes.Diagnostics
	for k, lv := range src {
		v, err := expandAsString(ctx, ep, lv)
		if err != nil {
			diags = append(diags, parsetypes.NewErrorDiagnostics(err, lv.GetLocation())...)
			continue
		}

		dst.Environment[k] = v
	}

	return diags
}

func expandAsString(ctx context.Context, ep expr.EvalParams, lv *LazyValue) (string, error) {
	raw, err := lv.Expand(ctx, ep)
	if err != nil {
		return "", err
	}

	return parsetypes.AnyToString(raw)
}

type MapArgsParams struct {
	// Values is source raw values.
	Values JobArgs

	// Spec is job inputs schema for type checking.
	Spec Inputs

	// EvalParams is evaluation params to expand expressions.
	EvalParams expr.EvalParams

	// EnvVars is environment variables map.
	// Used to resolve default input default value when it contains env var binding.
	EnvVars map[string]string
}

// MapArgsToInputs validates job arguments against inputs schema and returns expanded values.
//
// Main purpose of the function is to construct scope values for task execution.
func MapArgsToInputs(ctx context.Context, params MapArgsParams) (map[string]any, parsetypes.Diagnostics) {
	dst := make(map[string]any, len(params.Spec))
	diags := parsetypes.Diagnostics{}

	visited := map[string]struct{}{}
	for k, spec := range params.Spec {
		raw, ok := params.Values.Values[k]
		if ok {
			visited[k] = struct{}{}
			v, diag := mapLazyValue(
				ctx, raw,
				valueMapOpts{
					ep:       params.EvalParams,
					spec:     spec.Schema,
					allowNil: true, // will fallback to	defaults if empty
				},
			)

			if len(diag) > 0 {
				diags = append(diags, diag...)
				if diag.HasError() {
					continue
				}
			}

			if v != nil {
				dst[k] = v
				continue
			}
		}

		// Fallback - read defaults
		if spec.IsRequired() {
			diags = diags.Append(parameterRequiredDiagnostic(k, params.Values.Location))
			continue
		}

		// Env var takes precedence over explicit defaults
		// Other types of bindings don't matter as they're supported in CLI context.
		if spec.Binding != nil && spec.Binding.EnvVarName != "" && len(params.EnvVars) > 0 {
			k := spec.Binding.EnvVarName
			v, ok := params.EnvVars[k]
			if ok && v != "" {
				out, diag := mapEnvVar(spec, v)
				if len(diag) > 0 {
					diags = append(diags, diag...)
				}

				if !diags.HasError() {
					dst[k] = out
				}
				continue
			}
		}

		// Unpack explicit default value
		if spec.DefaultValue != nil {
			// TODO: build default value
			out, diag := mapLazyValue(
				ctx, &spec.DefaultValue.Value,
				valueMapOpts{
					ep:   params.EvalParams,
					spec: spec.Schema,
					note: "unable to expand default value",
				},
			)

			if len(diag) > 0 {
				diags = append(diags, diag...)
				if diag.HasError() {
					continue
				}
			}

			dst[k] = out
			continue
		}

		// Fallback for optionals
		if spec.Schema.Type == ValueTypeBool {
			dst[k] = false
			continue
		}

		if spec.Optional {
			dst[k] = NewZeroValue(spec.Schema.Type)
			continue
		}

		diags = diags.Append(parameterRequiredDiagnostic(k, params.Values.Location))
		continue
	}

	// Throw warnings about unknown passed fields
	for k, e := range params.Values.Values {
		if _, isConsumed := visited[k]; isConsumed {
			continue
		}

		diag := &parsetypes.Diagnostic{
			FileName: e.Location.FileName,
			Severity: parsetypes.DiagnosticSeverityWarning,
			Range:    e.Location.Range,
			Offset:   e.Location.Offset,
			Note:     "remove unnecessary parameter",
			Err:      fmt.Errorf("unknown parameter %q", k),
		}
		diags = diags.Append(diag)
	}

	return dst, diags
}

func mapEnvVar(def *InputDefinition, val string) (any, parsetypes.Diagnostics) {
	spec := def.Schema
	loc := def.Binding.Location.EnvVarName
	if loc == nil {
		loc = &def.Location
	}

	if spec.Type.IsList() {
		if spec.Items.Items.Type.IsComplex() {
			err := fmt.Errorf("binding environment variables to []%s is not supported", spec.Items.Items.Type)
			note := fmt.Sprintf("references environment variable %q", def.Binding.EnvVarName)
			return nil, newDiagnostic(loc, err, note)
		}

		parts := strings.Split(val, def.Binding.DelimiterOrDefault())
		out := make([]any, 0, len(parts))
		for _, part := range parts {
			v, err := spec.ParseValue(part)
			if err != nil {
				note := fmt.Sprintf("references environment variable %q", def.Binding.EnvVarName)
				return nil, newDiagnostic(loc, err, note)
			}

			out = append(out, v)
		}

		return out, nil
	}

	if spec.Type.IsComplex() {
		err := fmt.Errorf("binding environment variables to %s is not supported", spec.Type)
		note := fmt.Sprintf("references environment variable %q", def.Binding.EnvVarName)
		return nil, newDiagnostic(loc, err, note)
	}

	out, err := spec.ParseValue(val)
	if err != nil {
		note := fmt.Sprintf("references environment variable %q", def.Binding.EnvVarName)
		return nil, newDiagnostic(loc, err, note)
	}

	return out, nil
}

type valueMapOpts struct {
	ep       expr.EvalParams
	spec     TypeSchema
	note     string
	allowNil bool
}

func mapLazyValue(ctx context.Context, val *LazyValue, opts valueMapOpts) (any, parsetypes.Diagnostics) {
	raw, err := val.Expand(ctx, opts.ep)
	if err != nil {
		if diags, ok := parsetypes.IsDiagnosticsError(err); ok {
			return nil, diags
		}

		return nil, newDiagnostic(val.Location, err, opts.note)
	}

	return mapRawValue(raw, val.Location, opts)
}

func mapRawValue(raw any, loc *ReferenceLocation, opts valueMapOpts) (any, parsetypes.Diagnostics) {
	if t, ok := raw.(string); ok {
		out, err := opts.spec.ParseValue(t)
		if err != nil {
			return nil, newDiagnostic(loc, err, opts.note)
		}

		return out, nil
	}

	if isNilValue(raw) {
		if !opts.allowNil {
			return nil, newDiagnostic(loc, errors.New("value cannot be nil"), opts.note)
		}

		if !opts.spec.Type.IsComplex() {
			return NewZeroValue(opts.spec.Type), nil
		}

		return nil, nil
	}

	var (
		out any
		err error
	)

	switch opts.spec.Type {
	case ValueTypeString:
		out, err = parsetypes.AnyToString(raw)
	case ValueTypeBool:
		out, err = parsetypes.AnyToBool(raw)
	case ValueTypeInt:
		out, err = parsetypes.AnyToInt(raw)
	case ValueTypeFloat:
		out, err = parsetypes.AnyToFloat(raw)
	case ValueTypeDate:
		out, err = anyToDate(raw, opts.spec.DateFormatOrDefault())
	case ValueTypeDuration:
		out, err = parsetypes.AnyToDuration(raw)
	case ValueTypeList:
		out, err = anyToTypedList(raw, loc, opts)
	case ValueTypeDict:
		out, err = anyToTypedDict(raw, loc, opts)
	default:
		err = fmt.Errorf("unsupported value type: %s", opts.spec.Type)
	}

	if err != nil {
		return nil, newDiagnostic(loc, err, opts.note)
	}

	return out, nil
}

func isNilValue(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func anyToDate(v any, dateFormat string) (time.Time, error) {
	switch t := v.(type) {
	case time.Time:
		return t, nil
	case []byte:
		return time.Parse(dateFormat, string(t))
	default:
		return time.Time{}, fmt.Errorf("value of type %T cannot be converted to date", v)
	}
}

func anyToTypedList(v any, loc *ReferenceLocation, opts valueMapOpts) ([]any, error) {
	list, err := parsetypes.AnyToList(v)
	if err != nil || opts.spec.Items == nil {
		return list, err
	}

	out := make([]any, len(list))
	itemOpts := valueMapOpts{
		ep:       opts.ep,
		spec:     *opts.spec.Items,
		note:     opts.note,
		allowNil: true,
	}

	for i, item := range list {
		val, diags := mapRawValue(item, loc, itemOpts)
		if diags.HasError() {
			return nil, fmt.Errorf("invalid value at index %d: %w", i, diags[0].Err)
		}

		out[i] = val
	}

	return out, nil
}

func anyToTypedDict(v any, loc *ReferenceLocation, opts valueMapOpts) (map[string]any, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, fmt.Errorf("expected a dict, but got %s %#v", rv.Kind(), v)
	}

	out := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		key := iter.Key()
		if key.Kind() != reflect.String {
			return nil, fmt.Errorf("dict key type should be string, got %s", key.Kind())
		}

		out[key.String()] = iter.Value().Interface()
	}

	if opts.spec.Items == nil {
		return out, nil
	}

	itemOpts := valueMapOpts{
		ep:       opts.ep,
		spec:     *opts.spec.Items,
		note:     opts.note,
		allowNil: true,
	}

	for key, item := range out {
		val, diags := mapRawValue(item, loc, itemOpts)
		if diags.HasError() {
			return nil, fmt.Errorf("invalid value for key %q: %w", key, diags[0].Err)
		}

		out[key] = val
	}

	return out, nil
}

func newDiagnostic(loc *ReferenceLocation, err error, note string) parsetypes.Diagnostics {
	diag := &parsetypes.Diagnostic{
		FileName: loc.FileName,
		Severity: parsetypes.DiagnosticSeverityError,
		Range:    loc.Range,
		Offset:   loc.Offset,
		Err:      err,
		Note:     note,
	}

	return parsetypes.Diagnostics{diag}
}

func parameterRequiredDiagnostic(name string, argsLocation *ReferenceLocation) *parsetypes.Diagnostic {
	return &parsetypes.Diagnostic{
		FileName: argsLocation.FileName,
		Severity: parsetypes.DiagnosticSeverityError,
		Range:    argsLocation.Range,
		Offset:   argsLocation.Offset,
		Err:      fmt.Errorf("input parameter %q is required", name),
	}
}
