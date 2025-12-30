package argschema

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
	"github.com/go-gilbert/gilbert/pkg/expr"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type FieldVisitor[T any] interface {
	Name() string
	IsRequired() bool
	VisitLazyField(vp VisitParams, lz *manifest.LazyValue, dst *T) parsetypes.Diagnostics
}

type ValueVisitor[T any] interface {
	VisitValue(vp VisitParams, lz *manifest.LazyValue) (T, parsetypes.Diagnostics)
}

type ValueConverter[T any] interface {
	ConvertValue(vp VisitParams, v any) (T, error)
}

var (
	AnyFloat = &AnyValueVisitor[float64]{
		parseFunc: func(_ VisitParams, v any) (float64, error) {
			return parsetypes.AnyToFloat(v)
		},
	}

	AnyUint = &AnyValueVisitor[uint]{
		parseFunc: func(_ VisitParams, v any) (uint, error) {
			return parsetypes.AnyToUint(v)
		},
	}

	AnyInt = &AnyValueVisitor[int]{
		parseFunc: func(_ VisitParams, v any) (int, error) {
			r, err := parsetypes.AnyToInt(v)
			return int(r), err
		},
	}

	AnyBool = &AnyValueVisitor[bool]{
		parseFunc: func(_ VisitParams, v any) (bool, error) {
			return parsetypes.AnyToBool(v)
		},
	}

	AnyString = &AnyValueVisitor[string]{
		parseFunc: func(_ VisitParams, v any) (string, error) {
			return parsetypes.AnyToString(v)
		},
	}

	AnyDuration = &AnyValueVisitor[time.Duration]{
		parseFunc: func(_ VisitParams, v any) (time.Duration, error) {
			return parsetypes.AnyToDuration(v)
		},
	}
)

// MapStructure attempts to map value to a structure using mapstructure library.
func MapStructure[T any](hooks ...mapstructure.DecodeHookFunc) *AnyValueVisitor[T] {
	return &AnyValueVisitor[T]{
		parseFunc: func(_ VisitParams, v any) (T, error) {
			hooks = append(hooks, parsetypes.DecodeHookFunc)

			var res T
			dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
				Result:           &res,
				DecodeHook:       mapstructure.ComposeDecodeHookFunc(hooks...),
			})
			if err != nil {
				return res, err
			}

			err = dec.Decode(v)
			return res, err
		},
	}
}

// List maps any value to a list.
func List[T any](vis ValueConverter[T]) *AnyValueVisitor[[]T] {
	return &AnyValueVisitor[[]T]{
		parseFunc: func(vp VisitParams, v any) ([]T, error) {
			if v == nil {
				return nil, nil
			}

			ref := reflect.ValueOf(v)
			switch ref.Kind() {
			case reflect.Array, reflect.Slice:
				break
			case reflect.Map, reflect.Struct:
				return nil, errors.New("invalid type: expected a list but got object")
			default:
				return nil, fmt.Errorf("invalid type: expected a list but got %T", v)
			}

			c := ref.Len()
			out := make([]T, c)

			for i := range c {
				iface := ref.Index(i).Interface()
				tval, err := vis.ConvertValue(vp, iface)
				if err != nil {
					return nil, fmt.Errorf("invalid value at index %d: %w", i, err)
				}

				out[i] = tval
			}

			return out, nil
		},
	}
}

// OneOrMany is operator similar to [List], but can map a single value to a list.
func OneOrMany[T any](vis ValueConverter[T]) *AnyValueVisitor[[]T] {
	return &AnyValueVisitor[[]T]{
		parseFunc: func(vp VisitParams, v any) ([]T, error) {
			if v == nil {
				return nil, nil
			}

			ref := reflect.ValueOf(v)
			switch ref.Kind() {
			case reflect.Array, reflect.Slice:
				c := ref.Len()
				out := make([]T, c)

				for i := range c {
					iface := ref.Index(i).Interface()
					tval, err := vis.ConvertValue(vp, iface)
					if err != nil {
						return nil, fmt.Errorf("invalid value at index %d: %w", i, err)
					}

					out[i] = tval
				}

				return out, nil
			}

			one, err := vis.ConvertValue(vp, v)
			if err != nil {
				return nil, err
			}

			return []T{one}, nil
		},
	}
}

type AnyValueVisitor[T any] struct {
	parseFunc func(vp VisitParams, v any) (T, error)
}

func (vis AnyValueVisitor[T]) ConvertValue(vp VisitParams, rv any) (T, error) {
	return vis.parseFunc(vp, rv)
}

func (vis AnyValueVisitor[T]) VisitValue(vp VisitParams, lz *manifest.LazyValue) (v T, diags parsetypes.Diagnostics) {
	vp = vp.withLocation(lz.Location)
	raw, diags := expandVal(vp, lz)
	if diags.HasError() {
		return v, diags
	}

	val, err := vis.parseFunc(vp, raw)
	if err != nil {
		loc := lz.Location
		diags = append(diags, &parsetypes.Diagnostic{
			FileName: loc.FileName,
			Severity: parsetypes.SeverityFromError(err),
			Range:    loc.Range,
			Offset:   loc.Offset,
			Err:      err,
		})
	}

	return val, diags
}

type StructFieldVisitor[TObj any, TProp any] struct {
	name     string
	required bool
	visitor  ValueVisitor[TProp]
	setValue func(vp VisitParams, dst *TObj, val TProp) error
}

func Field[TObj, TProp any](name string, visitor ValueVisitor[TProp], setValue func(vp VisitParams, dst *TObj, val TProp) error) *StructFieldVisitor[TObj, TProp] {
	return &StructFieldVisitor[TObj, TProp]{
		name:     name,
		required: true,
		visitor:  visitor,
		setValue: setValue,
	}
}

func (vis *StructFieldVisitor[TObj, TProp]) Name() string {
	return vis.name
}

func (vis *StructFieldVisitor[TObj, TProp]) IsRequired() bool {
	return vis.required
}

func (vis *StructFieldVisitor[TObj, TProp]) Required() *StructFieldVisitor[TObj, TProp] {
	vis.required = true
	return vis
}

func (vis *StructFieldVisitor[TObj, TProp]) Optional() *StructFieldVisitor[TObj, TProp] {
	vis.required = false
	return vis
}

func (vis *StructFieldVisitor[TObj, TProp]) VisitLazyField(vp VisitParams, lz *manifest.LazyValue, dst *TObj) parsetypes.Diagnostics {
	v, diags := vis.visitor.VisitValue(vp, lz)
	if diags.HasError() {
		return diags
	}

	if err := vis.setValue(vp, dst, v); err != nil {
		return vp.withLocation(lz.Location).addDiagnostic(diags, err)
	}

	return diags
}

type VisitParams struct {
	Context        context.Context
	ParentLocation *manifest.ReferenceLocation
	EvalParams     expr.EvalParams
}

func (vp VisitParams) withLocation(loc *manifest.ReferenceLocation) VisitParams {
	vp.ParentLocation = loc
	return vp
}

func (vp VisitParams) addDiagnostic(diags parsetypes.Diagnostics, err error) parsetypes.Diagnostics {
	if vp.ParentLocation == nil {
		return diags
	}

	loc := vp.ParentLocation
	return append(diags, &parsetypes.Diagnostic{
		FileName: loc.FileName,
		Severity: parsetypes.DiagnosticSeverityError,
		Range:    loc.Range,
		Offset:   loc.Offset,
		Err:      err,
	})
}

type StructVisitor[T any] struct {
	fields      []FieldVisitor[T]
	knownFields map[string]struct{}
	constructor func(*T)
}

func Struct[T any](fields ...FieldVisitor[T]) *StructVisitor[T] {
	knownFields := map[string]struct{}{}
	for _, f := range fields {
		knownFields[f.Name()] = struct{}{}
	}

	return &StructVisitor[T]{
		fields:      fields,
		knownFields: knownFields,
	}
}

func (sv *StructVisitor[T]) Constructor(fn func(*T)) *StructVisitor[T] {
	sv.constructor = fn
	return sv
}

func (sv *StructVisitor[T]) VisitMap(vp VisitParams, kv map[string]*manifest.LazyValue, dst *T) parsetypes.Diagnostics {
	var (
		out   T
		diags parsetypes.Diagnostics
	)

	if sv.constructor != nil {
		sv.constructor(&out)
	}

	unvisitedFields := make(map[string]struct{}, len(kv))
	for k := range kv {
		unvisitedFields[k] = struct{}{}
	}

	for _, field := range sv.fields {
		k := field.Name()
		v, ok := kv[k]
		if !ok {
			if field.IsRequired() {
				diags = vp.addDiagnostic(diags, fmt.Errorf("missing required field %q", k))
			}

			continue
		}

		delete(unvisitedFields, k)
		fieldDiags := field.VisitLazyField(vp, v, &out)
		diags = append(diags, fieldDiags...)
		if fieldDiags.HasError() {
			return diags
		}
	}

	for prop := range unvisitedFields {
		loc := kv[prop].Location
		diags = append(diags, &parsetypes.Diagnostic{
			FileName: loc.FileName,
			Severity: parsetypes.DiagnosticSeverityWarning,
			Range:    loc.Range,
			Offset:   loc.Offset,
			Err:      fmt.Errorf("unknown field %q", prop),
			Note:     "remove unknown property",
		})
	}

	return diags
}

func expandVal(vp VisitParams, lzv *manifest.LazyValue) (any, parsetypes.Diagnostics) {
	v, err := lzv.Expand(vp.Context, vp.EvalParams)
	if err == nil {
		return v, nil
	}

	diags, ok := parsetypes.IsDiagnosticsError(err)
	if !ok {
		loc := lzv.Location
		diags = parsetypes.Diagnostics{
			&parsetypes.Diagnostic{
				FileName: loc.FileName,
				Severity: parsetypes.DiagnosticSeverityError,
				Range:    loc.Range,
				Offset:   loc.Offset,
				Err:      err,
			},
		}
	}

	return v, diags
}
