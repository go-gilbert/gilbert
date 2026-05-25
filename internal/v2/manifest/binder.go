package manifest

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

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
			diags.Append(parameterRequiredDiagnostic(k, params.Values.Location))
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

		diags.Append(parameterRequiredDiagnostic(k, params.Values.Location))
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
		diags.Append(diag)
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

	if t, ok := raw.(string); ok {
		out, err := opts.spec.ParseValue(t)
		if err != nil {
			return nil, newDiagnostic(val.Location, err, opts.note)
		}

		return out, nil
	}

	if !opts.allowNil {
		if raw == nil {
			return nil, newDiagnostic(val.Location, errors.New("value cannot be nil"), opts.note)
		}

		// TODO: handle iface with nil value
	}

	switch opts.spec.Type {
	// TODO: typecheck.
	// NOTE: some scalar values can be casted between.
	}

	return nil, nil
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
