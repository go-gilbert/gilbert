package manifest

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// DefaultDateFormat is standard date format for inputs.
//
// Default is RFC3339 which is aka ISO8601.
const DefaultDateFormat = time.RFC3339

type ValueType uint8

const (
	ValueTypeInvalid ValueType = iota
	ValueTypeString
	ValueTypeBool
	ValueTypeInt
	ValueTypeFloat
	ValueTypeList
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
	case ValueTypeList:
		return "list"
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
	case ValueTypeList:
		return "ValueTypeList"
	}

	return fmt.Sprint(v)
}

func (v ValueType) IsComplex() bool {
	return v == ValueTypeList
}

type ValueFormat uint8

const (
	ValueFormatInvalid ValueFormat = iota
	ValueFormatDuration
	ValueFormatDate
	ValueFormatURL
)

func (v ValueFormat) String() string {
	switch v {
	case ValueFormatDuration:
		return "duration"
	case ValueFormatDate:
		return "date"
	case ValueFormatURL:
		return "url"
	}

	return "<invalid>"
}

func (v ValueFormat) GoString() string {
	switch v {
	case ValueFormatDuration:
		return "ValueFormatDuration"
	case ValueFormatDate:
		return "ValueFormatDate"
	case ValueFormatURL:
		return "ValueFormatURL"
	}

	return fmt.Sprint(v)
}

func ParseValueFormat(format string) (ValueFormat, error) {
	switch format {
	case "duration":
		return ValueFormatDuration, nil
	case "date":
		return ValueFormatDate, nil
	case "url":
		return ValueFormatURL, nil
	}

	return ValueFormatInvalid, errors.New("invalid value format")
}

func ParseValueType(value string) (ValueType, error) {
	switch value {
	case "string":
		return ValueTypeString, nil
	case "int":
		return ValueTypeInt, nil
	case "bool":
		return ValueTypeBool, nil
	case "float":
		return ValueTypeFloat, nil
	case "list":
		return ValueTypeList, nil
	}

	return ValueTypeInvalid, errors.New("invalid value type")
}

type TypeSchema struct {
	Type       ValueType
	Format     ValueFormat
	DateFormat string
	Items      *TypeSchema
}

// ParseString parses input string using format specified in a type schema.
func (s TypeSchema) ParseString(val string) (any, error) {
	switch s.Format {
	case ValueFormatDate:
		dateFmt := s.DateFormatOrDefault()
		return time.Parse(dateFmt, val)
	case ValueFormatDuration:
		return time.ParseDuration(val)
	case ValueFormatURL:
		return url.Parse(val)
	case ValueFormatInvalid:
		return val, nil
	default:
		return nil, fmt.Errorf("unsupported value format: %q", s.Format)
	}
}

func (s TypeSchema) String() string {
	switch s.Type {
	case ValueTypeBool:
		return "boolean"
	case ValueTypeInt:
		return "int"
	case ValueTypeFloat:
		return "float"
	case ValueTypeList:
		itemsType := "nil"
		if s.Items != nil {
			itemsType = s.Items.String()
		}

		return "list[" + itemsType + "]"
	case ValueTypeString:
		break
	default:
		return "<invalid>"
	}

	switch s.Format {
	case ValueFormatDate:
		return "date"
	case ValueFormatDuration:
		return "duration"
	case ValueFormatURL:
		return "url"
	case ValueFormatInvalid:
		return "string"
	default:
		return "<invalid>"
	}
}

func (s TypeSchema) DateFormatOrDefault() string {
	if s.DateFormat == "" {
		return DefaultDateFormat
	}

	return s.DateFormat
}
