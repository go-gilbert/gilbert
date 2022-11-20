package spec

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
	"reflect"
	"time"
)

var (
	typeRefCapsule = cty.Capsule("type", reflect.TypeOf(cty.Type{}))
	timeCapsule    = cty.Capsule("type", reflect.TypeOf(time.Time{}))

	types = map[string]cty.Value{
		"number":  exportCtyType(&cty.Number),
		"string":  exportCtyType(&cty.String),
		"boolean": exportCtyType(&cty.Bool),
		"object":  exportCtyType(&cty.EmptyObject),
		"date":    exportCtyType(&timeCapsule),
	}

	compoundTypes = map[string]function.Function{
		"array": genericTypeConstructor(1, func(args []cty.Type) cty.Type {
			return cty.List(args[0])
		}),
		"map": genericTypeConstructor(1, func(args []cty.Type) cty.Type {
			return cty.Map(args[1])
		}),
		"typeof": function.New(&function.Spec{
			Description: "Obtain a type of value",
			Params: []function.Parameter{
				{Type: typeRefCapsule},
			},
			Type: function.StaticReturnType(typeRefCapsule),
			Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
				typ := args[0].Type()
				return gocty.ToCtyValue(typ, typeRefCapsule)
			},
		}),
	}

	typesContext = &hcl.EvalContext{
		Variables: types,
		Functions: compoundTypes,
	}
)

func ctyTypeFromExpression(expr hcl.Expression) (cty.Type, hcl.Diagnostics) {
	val, err := expr.Value(typesContext)
	if err != nil {
		return cty.NilType, err
	}

	var typ cty.Type
	if err := gocty.FromCtyValue(val, &typ); err != nil {
		return typ, newDiagnosticError(expr.Range(),
			"attribute can only accept types, not values (%s)", err)
	}

	return typ, nil
}

func exportCtyType(t *cty.Type) cty.Value {
	v, err := gocty.ToCtyValue(t, typeRefCapsule)
	if err != nil {
		panic(fmt.Sprintf("exportCtyType(%s): %s", t.FriendlyName(), err))
	}
	return v
}

type typeConstructor = func(args []cty.Type) cty.Type

func genericTypeConstructor(argCount int, constructor typeConstructor) function.Function {
	s := &function.Spec{
		Params: make([]function.Parameter, argCount),
		Type:   function.StaticReturnType(typeRefCapsule),
		Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
			types := make([]cty.Type, len(args))
			for i, typeVal := range args {
				var dst cty.Type
				if err := gocty.FromCtyValue(typeVal, &dst); err != nil {
					return cty.NilVal, fmt.Errorf(
						"failed to obtain type spec from argument %d: %w",
						i+1, err)
				}

				types[i] = dst
			}

			t := constructor(types)
			return gocty.ToCtyValue(t, typeRefCapsule)
		},
	}

	return function.New(s)
}
