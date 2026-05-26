package inputflag

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-gilbert/gilbert/internal/manifest"
)

type ValueDecoderFunc = func(val string) (any, error)

func getValueDecoder(typ manifest.TypeSchema) (ValueDecoderFunc, error) {
	// TODO: move this into "manifest"?
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
	case manifest.ValueTypeDate:
		parseFn = func(val string) (any, error) {
			if val == "" {
				return nil, nil
			}

			return time.Parse(typ.DateFormatOrDefault(), val)
		}
	case manifest.ValueTypeDuration:
		parseFn = func(val string) (any, error) {
			if val == "" {
				return nil, nil
			}

			return time.ParseDuration(val)
		}
	case manifest.ValueTypeString:
		parseFn = func(val string) (any, error) {
			return val, nil
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

func getStringFormatter(typ manifest.ValueType, dateFormat string) func(any) string {
	if typ == manifest.ValueTypeDate {
		return func(v any) string {
			if dt, ok := v.(time.Time); ok {
				return strconv.Quote(dt.Format(dateFormat))
			}

			return fmt.Sprint(v)
		}
	}

	return mapToString
}
