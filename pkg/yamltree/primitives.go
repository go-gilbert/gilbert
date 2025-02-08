package yamltree

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"golang.org/x/exp/constraints"
)

type uintVisitor[T constraints.Unsigned] struct {
}

func (_ uintVisitor[T]) VisitItem(_ context.Context, fi FileInfo, node ast.Node) (T, parsetypes.Diagnostics) {
	switch t := node.Type(); t {
	case ast.NullType:
		return 0, nil
	case ast.IntegerType:
		break
	default:
		return 0, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node, fmt.Errorf("value should be unsigned integer, got %s", t)),
		}
	}

	v, err := strconv.ParseUint(node.GetToken().Value, 0, 64)
	if err != nil {
		return 0, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node, err),
		}
	}

	return T(v), nil
}

type intVisitor[T constraints.Signed] struct{}

func (_ intVisitor[T]) VisitItem(_ context.Context, fi FileInfo, node ast.Node) (T, parsetypes.Diagnostics) {
	switch t := node.Type(); t {
	case ast.NullType:
		return 0, nil
	case ast.IntegerType:
		break
	default:
		return 0, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node, fmt.Errorf("value should be integer, got %s", t)),
		}
	}

	v, err := strconv.ParseInt(node.GetToken().Value, 0, 64)
	if err != nil {
		return 0, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node, err),
		}
	}

	return T(v), nil
}

type floatVisitor[T constraints.Float] struct{}

func (_ floatVisitor[T]) VisitItem(_ context.Context, fi FileInfo, node ast.Node) (T, parsetypes.Diagnostics) {
	var val string
	switch t := node.Type(); t {
	case ast.IntegerType, ast.FloatType:
		val = node.GetToken().Value
	case ast.NullType:
		return 0, nil
	default:
		return 0, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node, fmt.Errorf("value should be float, got %s", t)),
		}
	}

	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node, err),
		}
	}

	return T(v), nil
}

type boolVisitor struct{}

func (_ boolVisitor) VisitItem(_ context.Context, fi FileInfo, node ast.Node) (bool, parsetypes.Diagnostics) {
	if IsNullNode(node) {
		return false, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node,
				errors.New("missing value"),
			),
		}
	}

	v, ok := node.(*ast.BoolNode)
	if !ok {
		return false, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node,
				fmt.Errorf("expected boolean value, got %s", node.Type()),
			),
		}
	}

	return v.Value, nil
}

type stringVisitor struct {
	strict bool
}

func (v stringVisitor) VisitItem(_ context.Context, fi FileInfo, node ast.Node) (string, parsetypes.Diagnostics) {
	typ := node.Type()
	switch typ {
	case ast.NullType:
		return "", nil
	case ast.StringType:
		return node.GetToken().Value, nil
	case ast.BoolType,
		ast.IntegerType,
		ast.FloatType:
		if !v.strict {
			return node.GetToken().Value, nil
		}
	default:
		break
	}

	return "", parsetypes.Diagnostics{
		newErrDiagnosticFromNode(fi.FileName, node,
			fmt.Errorf("value of type %s cannot be converted to a string", typ),
		),
	}
}

type unmarshalVisitor[T any] struct{}

func (v unmarshalVisitor[T]) VisitItem(ctx context.Context, fi FileInfo, node ast.Node) (T, parsetypes.Diagnostics) {
	var dst T

	err := yaml.NewDecoder(nil).DecodeFromNodeContext(ctx, node, &dst)
	if err != nil {
		// TODO: map yaml to diagnostics
		return dst, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(fi.FileName, node, err),
		}
	}

	return dst, nil
}
