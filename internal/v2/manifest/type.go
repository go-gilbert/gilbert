package manifest

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// DefaultDateFormat is standard date format for inputs.
//
// Default is RFC3339 which is aka ISO8601.
const DefaultDateFormat = time.RFC3339

// ValueType is value data type.
type ValueType uint8

const (
	ValueTypeInvalid ValueType = iota
	ValueTypeString
	ValueTypeBool
	ValueTypeInt
	ValueTypeFloat
	ValueTypeDate
	ValueTypeDuration
	ValueTypeList
	ValueTypeDict
)

func (v ValueType) String() string {
	switch v {
	case ValueTypeString:
		return "string"
	case ValueTypeBool:
		return "bool"
	case ValueTypeInt:
		return "int"
	case ValueTypeFloat:
		return "float"
	case ValueTypeDate:
		return "date"
	case ValueTypeDuration:
		return "duration"
	case ValueTypeList:
		return "list"
	case ValueTypeDict:
		return "dict"
	}

	return "<invalid>"
}

func (v ValueType) GoString() string {
	switch v {
	case ValueTypeString:
		return "ValueTypeString"
	case ValueTypeBool:
		return "ValueTypeBool"
	case ValueTypeInt:
		return "ValueTypeInt"
	case ValueTypeFloat:
		return "ValueTypeFloat"
	case ValueTypeDate:
		return "ValueTypeDate"
	case ValueTypeDuration:
		return "ValueTypeDuration"
	case ValueTypeList:
		return "ValueTypeList"
	case ValueTypeDict:
		return "ValueTypeDict"
	}

	return fmt.Sprint(v)
}

func (v ValueType) IsList() bool {
	return v == ValueTypeList
}

func (v ValueType) IsComplex() bool {
	switch v {
	case ValueTypeDict, ValueTypeList:
		return true
	default:
		return false
	}
}

// NewZeroValue returns a new empty value of a given type.
//
// Returns nil for complex types (maps, arrays, etc).
func NewZeroValue(t ValueType) any {
	switch t {
	case ValueTypeString:
		return ""
	case ValueTypeDate:
		return time.Now()
	case ValueTypeDuration:
		return time.Duration(0)
	case ValueTypeFloat:
		return float64(0)
	case ValueTypeInt:
		return int64(0)
	case ValueTypeBool:
		return false
	default:
		return nil
	}
}

// ParseValueType parses value kind from string representation.
func ParseValueType(value string) (ValueType, error) {
	switch value {
	case "string":
		return ValueTypeString, nil
	case "int":
		return ValueTypeInt, nil
	case "bool":
		return ValueTypeBool, nil
	case "time":
		return ValueTypeDate, nil
	case "duration":
		return ValueTypeDuration, nil
	case "float":
		return ValueTypeFloat, nil
	case "list":
		return ValueTypeList, nil
	}

	return ValueTypeInvalid, errors.New("invalid value type")
}

type TypeSchema struct {
	Type       ValueType
	DateFormat string
	Items      *TypeSchema
}

// ParseValue parses value from string representation using type schema.
//
// This method doesn't support complex values such as arrays or maps.
func (s TypeSchema) ParseValue(val string) (any, error) {
	var (
		parsedValue any
		err         error
	)

	switch t := s.Type; t {
	case ValueTypeString:
		return val, nil
	case ValueTypeInt:
		parsedValue, err = strconv.ParseInt(val, 10, 64)
	case ValueTypeFloat:
		parsedValue, err = strconv.ParseFloat(val, 64)
	case ValueTypeBool:
		parsedValue, err = strconv.ParseBool(val)
	case ValueTypeDate:
		parsedValue, err = time.Parse(s.DateFormatOrDefault(), val)
	case ValueTypeDuration:
		parsedValue, err = time.ParseDuration(val)
	default:
		return nil, fmt.Errorf("cannot parse string %q as %s", val, t)
	}

	if err != nil {
		return nil, err
	}

	return parsedValue, nil
}

func (s TypeSchema) String() string {
	switch s.Type {
	case ValueTypeBool:
		return "boolean"
	case ValueTypeInt:
		return "int"
	case ValueTypeFloat:
		return "float"
	case ValueTypeDate:
		return "date"
	case ValueTypeDuration:
		return "duration"
	case ValueTypeString:
		return "string"
	case ValueTypeList:
		itemsType := "nil"
		if s.Items != nil {
			itemsType = s.Items.String()
		}

		return "list[" + itemsType + "]"
	}

	return "<invalid>"
}

func (s TypeSchema) DateFormatOrDefault() string {
	if s.DateFormat == "" {
		return DefaultDateFormat
	}

	return s.DateFormat
}
