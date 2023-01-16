package ctyutil

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/go-gilbert/gilbert/v2/internal/util/collections"
	"github.com/zclconf/go-cty/cty"
)

// ErrNotPrimitive occurs when passed cty.Type or cty.Value is not primitive.
var ErrNotPrimitive = errors.New("value type is not primitive")

type StringDecodeError struct {
	val string
	typ cty.Type
	err error
}

// wrapStringDecodeError wraps error as string decode error
func wrapStringDecodeError(err error, typ cty.Type, val string) *StringDecodeError {
	return &StringDecodeError{
		val: val,
		typ: typ,
		err: err,
	}
}

func (s StringDecodeError) Error() string {
	return fmt.Sprintf("cannot decode string %q as %s", s.val, TypeToString(s.typ))
}

func (s StringDecodeError) Unwrap() error {
	return s.err
}

// ValueFromString decodes string to a new cty.Value.
//
// Supports only primitive types.
func ValueFromString(typ cty.Type, str string) (cty.Value, error) {
	if !typ.IsPrimitiveType() {
		return cty.NilVal, ErrNotPrimitive
	}

	switch typ {
	case cty.Bool:
		v, err := strconv.ParseBool(str)
		if err != nil {
			return cty.NilVal, wrapStringDecodeError(err, typ, str)
		}

		return cty.BoolVal(v), nil
	case cty.String:
		return cty.StringVal(str), nil
	case cty.Number:
		num := new(big.Float)
		_, _, err := num.Parse(str, 0)
		if err != nil {
			return cty.NilVal, wrapStringDecodeError(err, typ, str)
		}

		return cty.NumberVal(num), nil
	}

	return cty.NilVal, fmt.Errorf("cannot decode %q: %w", str, ErrNotPrimitive)
}

// TypeToString returns human-readable string representation of cty.Type.
func TypeToString(typ cty.Type) string {
	switch typ {
	case cty.String:
		return "string"
	case cty.Number:
		return "number"
	case cty.Bool:
		return "boolean"
	case cty.NilType:
		return "nil"
	}

	sb := strings.Builder{}
	if typ.IsTupleType() {
		// HCL treats arrays as tuples
		sb.WriteString("array(")
		typesToString(&sb, typ.TupleElementTypes()...)
		sb.WriteRune(')')
		return sb.String()
	}

	if listElemType := typ.ListElementType(); listElemType != nil {
		sb.WriteString("array(")
		typesToString(&sb, *listElemType)
		sb.WriteRune(')')
		return sb.String()
	}

	return typ.GoString()
}

func typesToString(sb *strings.Builder, elemTypes ...cty.Type) {
	uniqueTypes := make(collections.Set[string], len(elemTypes))
	for _, t := range elemTypes {
		uniqueTypes.Append(TypeToString(t))
	}

	i := 0
	for typ := range uniqueTypes {
		sb.WriteString(typ)
		if i < len(uniqueTypes)-1 {
			sb.WriteRune('|')
		}
		i++
	}
}
