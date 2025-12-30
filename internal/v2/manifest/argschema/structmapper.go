package argschema

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"

	"github.com/go-gilbert/gilbert/internal/v2/engine"
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

type optionalVisitor[T any] struct {
	vis ValueVisitor[T]
}

func (op optionalVisitor[T]) ConvertValue(vp VisitParams, rv any) (opt Option[T], err error) {
	converter, ok := op.vis.(ValueConverter[T])
	if !ok {
		return opt, fmt.Errorf("cannot convert value into %T: not supported", opt)
	}

	if rv == nil {
		return opt, nil
	}

	v, err := converter.ConvertValue(vp, rv)
	if err != nil {
		return opt, err
	}

	return Option[T]{
		ok:  true,
		val: v,
	}, nil
}

func (op optionalVisitor[T]) VisitValue(vp VisitParams, lz *manifest.LazyValue) (opt Option[T], diags parsetypes.Diagnostics) {
	if lz == nil {
		return opt, nil
	}

	v, diags := op.vis.VisitValue(vp, lz)
	if diags.HasError() {
		return opt, diags
	}

	return Option[T]{
		ok:  true,
		val: v,
	}, diags
}

// Optional returns visitor to read [Option] values.
func Optional[T any](vis ValueVisitor[T]) ValueVisitor[Option[T]] {
	return &optionalVisitor[T]{
		vis: vis,
	}
}

// Dict map value into a dictionary with string keys.
func Dict[T any](valType ValueConverter[T]) *AnyValueVisitor[map[string]T] {
	return &AnyValueVisitor[map[string]T]{
		parseFunc: func(vp VisitParams, v any) (map[string]T, error) {
			r := reflect.ValueOf(v)
			if k := r.Kind(); k != reflect.Map {
				return nil, fmt.Errorf("invalid type: expected dict but got %s", k)
			}

			out := make(map[string]T)
			iter := r.MapRange()
			for iter.Next() {
				rk := iter.Key()
				if rk.Kind() != reflect.String {
					return nil, fmt.Errorf("dict key type should be string, got %s", rk.Kind())
				}

				k := rk.String()
				rv := iter.Value().Interface()
				if rv == nil {
					// Skip empty values. This is used to allow optional values in expressions.
					continue
				}

				tval, err := valType.ConvertValue(vp, rv)
				if err != nil {
					return nil, fmt.Errorf("invalid value for key %q: %w", k, tval)
				}

				out[k] = tval
			}

			return out, nil
		},
	}
}

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
			out := make([]T, 0, c)

			for i := range c {
				iface := ref.Index(i).Interface()
				if iface == nil {
					// Skip empty values. This is used to allow optional values in expressions.
					continue
				}

				tval, err := vis.ConvertValue(vp, iface)
				if err != nil {
					return nil, fmt.Errorf("invalid value at index %d: %w", i, err)
				}

				out = append(out, tval)
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
				out := make([]T, 0, c)

				for i := range c {
					iface := ref.Index(i).Interface()

					// Skip empty values. This is used to allow optional values in expressions.
					if iface == nil {
						continue
					}

					tval, err := vis.ConvertValue(vp, iface)
					if err != nil {
						return nil, fmt.Errorf("invalid value at index %d: %w", i, err)
					}

					out = append(out, tval)
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

func Field[TObj, TProp any](name string, visitor ValueVisitor[TProp], setValue func(vp VisitParams, dst *TObj, val TProp) error) *StructFieldVisitor[TObj, TProp] {
	return &StructFieldVisitor[TObj, TProp]{
		name:     name,
		visitor:  visitor,
		setValue: setValue,
	}
}

func StringField[TObj any](name string, setValue func(dst *TObj, val string) error) *StructFieldVisitor[TObj, string] {
	return &StructFieldVisitor[TObj, string]{
		name:    name,
		visitor: AnyString,
		setValue: func(_ VisitParams, dst *TObj, v string) error {
			return setValue(dst, v)
		},
	}
}

func BoolField[TObj any](name string, setValue func(dst *TObj, val bool) error) *StructFieldVisitor[TObj, bool] {
	return &StructFieldVisitor[TObj, bool]{
		name:    name,
		visitor: AnyBool,
		setValue: func(_ VisitParams, dst *TObj, v bool) error {
			return setValue(dst, v)
		},
	}
}

func ListField[TObj, TProp any](name string, vdec ValueConverter[TProp], setValue func(dst *TObj, val []TProp) error) *StructFieldVisitor[TObj, []TProp] {
	return &StructFieldVisitor[TObj, []TProp]{
		name:    name,
		visitor: OneOrMany(vdec),
		setValue: func(_ VisitParams, dst *TObj, v []TProp) error {
			return setValue(dst, v)
		},
	}
}

type StructFieldVisitor[TObj any, TProp any] struct {
	name     string
	required bool
	visitor  ValueVisitor[TProp]
	setValue func(vp VisitParams, dst *TObj, val TProp) error
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

type EmbeddedStructVisitor[TParent, TEmbedded any] struct {
	vis         *StructVisitor[TEmbedded]
	getEmbedded func(parent *TParent) *TEmbedded
}

func (ev EmbeddedStructVisitor[TParent, TEmbedded]) ReadMap(vp VisitParams, unvisitedFields stringSet, parent *TParent, kv map[string]*manifest.LazyValue) parsetypes.Diagnostics {
	embedded := ev.getEmbedded(parent)
	return ev.vis.readMap(vp, unvisitedFields, embedded, kv)
}

// Embedded creates visitor for embedded structs.
//
// Visitor can be passed into [StructVisitor.Embedded].
func Embedded[TParent, TEmbedded any](getter func(parent *TParent) *TEmbedded, fields ...FieldVisitor[TEmbedded]) *EmbeddedStructVisitor[TParent, TEmbedded] {
	return &EmbeddedStructVisitor[TParent, TEmbedded]{
		vis:         Struct(fields...),
		getEmbedded: getter,
	}
}

type EmbeddedVisitor[T any] interface {
	ReadMap(vp VisitParams, unvisitedFields stringSet, parent *T, kv map[string]*manifest.LazyValue) parsetypes.Diagnostics
}

type stringSet = map[string]struct{}

type StructVisitor[T any] struct {
	fields      []FieldVisitor[T]
	embeds      []EmbeddedVisitor[T]
	constructor func(*T)
}

func Struct[T any](fields ...FieldVisitor[T]) *StructVisitor[T] {
	return &StructVisitor[T]{
		fields: fields,
	}
}

func (sv *StructVisitor[T]) Embedded(v EmbeddedVisitor[T]) *StructVisitor[T] {
	sv.embeds = append(sv.embeds, v)
	return sv
}

func (sv *StructVisitor[T]) Constructor(fn func(*T)) *StructVisitor[T] {
	sv.constructor = fn
	return sv
}

func (sv *StructVisitor[T]) VisitMap(vp VisitParams, kv map[string]*manifest.LazyValue) (T, parsetypes.Diagnostics) {
	var out T

	if sv.constructor != nil {
		sv.constructor(&out)
	}

	unvisitedFields := make(stringSet, len(kv))
	for k := range kv {
		unvisitedFields[k] = struct{}{}
	}

	diags := sv.readMap(vp, unvisitedFields, &out, kv)
	for _, vis := range sv.embeds {
		diags = append(diags, vis.ReadMap(vp, unvisitedFields, &out, kv)...)
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

	return out, diags
}

func (sv *StructVisitor[T]) readMap(vp VisitParams, unvisitedFields stringSet, dst *T, kv map[string]*manifest.LazyValue) (diags parsetypes.Diagnostics) {
	for _, field := range sv.fields {
		k := field.Name()
		v, ok := kv[k]
		if !ok {
			if field.IsRequired() {
				diags = vp.addDiagnostic(diags, fmt.Errorf("missing required field %q", k))
			}

			continue
		}

		if v == nil {
			// Skip empty
			continue
		}

		delete(unvisitedFields, k)
		fieldDiags := field.VisitLazyField(vp, v, dst)
		diags = append(diags, fieldDiags...)
		if fieldDiags.HasError() {
			return diags
		}
	}

	return diags
}

func MapArgsToStruct[T any](ctx context.Context, ap engine.ActionParams, schema *StructVisitor[T]) (T, parsetypes.Diagnostics) {
	vp := VisitParams{
		Context:        ctx,
		ParentLocation: ap.Args.Location,
		EvalParams:     ap.EvalParams,
	}

	return schema.VisitMap(vp, ap.Args.Values)
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
