package inputflag

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-gilbert/gilbert/internal/v2/manifest"
)

func decodeValueWithSchema(val string, typ manifest.TypeSchema) (any, error) {
	if val == "" {
		return nil, nil
	}

	var (
		parsedValue any
		err         error
	)
	switch t := typ.Type; t {
	case manifest.ValueTypeInt:
		parsedValue, err = strconv.ParseInt(val, 10, 64)
	case manifest.ValueTypeFloat:
		parsedValue, err = strconv.ParseFloat(val, 64)
	case manifest.ValueTypeBool:
		parsedValue, err = strconv.ParseBool(val)
	case manifest.ValueTypeString:
		parsedValue, err = typ.ParseString(val)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %q as %s: %w", val, typ.Format, err)
		}

	default:
		// TODO: support list of scalars
		return nil, fmt.Errorf("cannot parse string %q as %s", val, t)
	}

	if err != nil {
		return nil, err
	}

	//i.flagCtx.dstScope.Inputs[i.inputDef.Name] = parsedValue
	return parsedValue, nil
}

type ValueDecoderFunc = func(val string) (any, error)

func getValueDecoder(typ manifest.TypeSchema) (ValueDecoderFunc, error) {
	var parseFn ValueDecoderFunc
	switch t := typ.Type; t {
	case manifest.ValueTypeInt:
		parseFn = func(val string) (any, error) {
			if val == "" {
				return nil, nil
			}

			return strconv.ParseInt(val, 10, 64)
		}
	case manifest.ValueTypeFloat:
		parseFn = func(val string) (any, error) {
			if val == "" {
				return nil, nil
			}

			return strconv.ParseFloat(val, 64)
		}
	case manifest.ValueTypeBool:
		parseFn = func(val string) (any, error) {
			if val == "" {
				return nil, nil
			}

			return strconv.ParseBool(val)
		}
	case manifest.ValueTypeString:
		parseFn = func(val string) (any, error) {
			if val == "" {
				return nil, nil
			}

			return typ.ParseString(val)
		}
	default:
		return nil, fmt.Errorf("unsupported input list item type: %s", typ)
	}

	return parseFn, nil
}

func mapToString(v any) string {
	switch t := v.(type) {
	case fmt.Stringer:
		return strconv.Quote(t.String())
	case string:
		return strconv.Quote(t)
	default:
		return fmt.Sprint(v)
	}
}

func getStringFormatter(typ manifest.TypeSchema, format manifest.ValueFormat) func(any) string {
	if format == manifest.ValueFormatDate {
		return func(v any) string {
			if dt, ok := v.(time.Time); ok {
				return strconv.Quote(dt.Format(typ.DateFormatOrDefault()))
			}

			return fmt.Sprint(v)
		}
	}

	return mapToString
}
