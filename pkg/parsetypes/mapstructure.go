package parsetypes

import (
	"reflect"
	"time"
)

var timeDurationTyp = reflect.TypeOf(time.Duration(0))

// DecodeHookFunc is value decode hook for mapstructure library.
//
// Implements the same best effort conversion logic as "AnyToX" helpers.
func DecodeHookFunc(f reflect.Type, t reflect.Type, v any) (any, error) {
	// TODO: add rest of cases
	switch t {
	case timeDurationTyp:
		return AnyToDuration(v)
	}

	return v, nil
}
