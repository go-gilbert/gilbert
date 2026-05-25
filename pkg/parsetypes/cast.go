package parsetypes

import (
	"fmt"
	"iter"
	"reflect"
	"slices"
	"strconv"
	"time"
)

func AnyToString(v any) (string, error) {
	switch t := v.(type) {
	case []byte:
		return string(t), nil
	case string:
		return t, nil
	case bool:
		return strconv.FormatBool(t), nil
	case uint8:
		return strconv.FormatUint(uint64(t), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(t), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(t), 10), nil
	case uint64:
		return strconv.FormatUint(t, 10), nil
	case int:
		return strconv.Itoa(t), nil
	case int8:
		return strconv.FormatInt(int64(t), 10), nil
	case int16:
		return strconv.FormatInt(int64(t), 10), nil
	case int32:
		return strconv.FormatInt(int64(t), 10), nil
	case int64:
		return strconv.FormatInt(t, 10), nil
	case float32:
		return strconv.FormatFloat(float64(t), 'g', -1, 32), nil
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64), nil
	default:
		return "", fmt.Errorf("value of type %T cannot be converted to string", v)
	}
}

func AnyToBool(v any) (bool, error) {
	if v == nil {
		return false, nil
	}

	// TODO: allow only 0 or 1 for bool.
	switch t := v.(type) {
	case bool:
		return t, nil
	case int:
		return t != 0, nil
	case int8:
		return t != 0, nil
	case int16:
		return t != 0, nil
	case int32:
		return t != 0, nil
	case int64:
		return t != 0, nil
	case uint:
		return t != 0, nil
	case uint8:
		return t != 0, nil
	case uint16:
		return t != 0, nil
	case uint32:
		return t != 0, nil
	case uint64:
		return t != 0, nil
	case float32:
		return t != 0, nil
	case float64:
		return t != 0, nil
	case string:
		return boolFromStr(t)
	case []byte:
		return boolFromStr(string(t))
	default:
		return false, fmt.Errorf("value of type %T cannot be converted to bool", t)
	}
}

func boolFromStr(v string) (bool, error) {
	r, err := strconv.ParseBool(v)
	if err != nil {
		// intercept Go's error and return human-readable error
		err = fmt.Errorf("cannot parse string %q as boolean", v)
	}

	return r, err
}

func AnyToUint(v any) (uint, error) {
	u := 0
	switch t := v.(type) {
	case int:
		u = int(t)
	case int8:
		u = int(t)
	case int16:
		u = int(t)
	case int32:
		u = int(t)
	case int64:
		u = int(t)
	case uint:
		return uint(t), nil
	case uint8:
		return uint(t), nil
	case uint16:
		return uint(t), nil
	case uint32:
		return uint(t), nil
	case uint64:
		return uint(t), nil
	case float32:
		u = int(t)
	case float64:
		u = int(t)
	default:
		return 0, fmt.Errorf("value of type %T cannot be converted to int", t)
	}

	if u < 0 {
		return 0, fmt.Errorf("value %v should be equal or greater than zero", u)
	}

	return uint(u), nil
}

func AnyToInt(v any) (int64, error) {
	switch t := v.(type) {
	case int:
		return int64(t), nil
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case int64:
		return t, nil
	case uint:
		return int64(t), nil
	case uint8:
		return int64(t), nil
	case uint16:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		return int64(t), nil
	case float32:
		return int64(t), nil
	case float64:
		return int64(t), nil
	default:
		return 0, fmt.Errorf("value of type %T cannot be converted to int", t)
	}
}

func AnyToFloat(v any) (float64, error) {
	switch t := v.(type) {
	case int8:
		return float64(t), nil
	case int16:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case uint8:
		return float64(t), nil
	case uint16:
		return float64(t), nil
	case uint32:
		return float64(t), nil
	case uint64:
		return float64(t), nil
	case float32:
		return float64(t), nil
	case float64:
		return t, nil
	default:
		return 0, fmt.Errorf("value of type %T cannot be converted to float", t)
	}
}

func AnyToDuration(v any) (time.Duration, error) {
	var u time.Duration
	switch t := v.(type) {
	case time.Duration:
		return t, nil
	case string:
		dur, err := time.ParseDuration(t)
		if err == nil {
			return dur, nil
		}

		// Handle case when number is returned as string from expression
		n, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse string as time duration or number of miliseconds")
		}

		u = time.Duration(n)
	case int:
		u = time.Duration(t)
	case int8:
		u = time.Duration(t)
	case int16:
		u = time.Duration(t)
	case int32:
		u = time.Duration(t)
	case int64:
		u = time.Duration(t)
	case uint:
		u = time.Duration(t)
	case uint8:
		return time.Duration(t), nil
	case uint16:
		return time.Duration(t), nil
	case uint32:
		return time.Duration(t), nil
	case uint64:
		return time.Duration(t), nil
	default:
		return 0, fmt.Errorf("expected duration string or number of miliseconds, but got %T", t)
	}

	if u < 0 {
		return 0, fmt.Errorf("number of miliseconds should be >= 0, got: %v", u)
	}

	return u * time.Millisecond, nil
}

func AnyToList(v any) ([]any, error) {
	if v == nil {
		return nil, nil
	}

	if l, ok := v.([]any); ok {
		return l, nil
	}

	r := reflect.ValueOf(v)
	switch k := r.Kind(); k {
	case reflect.Slice, reflect.Array:
		break
	default:
		return nil, fmt.Errorf("expected a list, but got %s %#v", k, v)
	}

	count := r.Len()
	dst := make([]any, count)
	for i := range count {
		dst[i] = r.Index(i).Interface()
	}

	return dst, nil
}

func IterAny(v any) (iter.Seq2[int, any], int, error) {
	if l, ok := v.([]any); ok {
		return slices.All(l), len(l), nil
	}

	r := reflect.ValueOf(v)
	switch k := r.Kind(); k {
	case reflect.Slice, reflect.Array:
		break
	default:
		return nil, 0, fmt.Errorf("expected a list, but got %s %#v", k, v)
	}

	count := r.Len()
	return func(yield func(int, any) bool) {
		for i := range count {
			v := r.Index(i).Interface()
			yield(i, v)
		}
	}, count, nil
}
