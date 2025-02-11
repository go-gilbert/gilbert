package manifest2

import (
	"errors"
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
	Location   ReferenceLocation
	Type       ValueType
	Format     ValueFormat
	DateFormat string
	Items      *TypeSchema
}

type InputBinding struct {
	EnvVarName string
}

type InputDefinition struct {
	TypeSchema
	Binding      *InputBinding
	DefaultValue any
}
