package yamltree

import (
	"context"
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type listVisitor[T any] struct {
	listItemVisitor ValueVisitor[T]
}

func (v listVisitor[T]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) ([]T, parsetypes.Diagnostics) {
	if IsNullNode(node) {
		return nil, nil
	}

	var diags parsetypes.Diagnostics
	sq, ok := node.(*ast.SequenceNode)
	if !ok {
		diags = append(diags,
			NewErrDiagnosticFromNode(
				opts.FileName, node,
				errors.New("node should be a list"),
			),
		)
		return nil, diags
	}

	out := make([]T, 0, len(sq.Values))
	for _, n := range sq.Values {
		if IsNullNode(n) {
			diags = append(diags,
				NewErrDiagnosticFromNode(
					opts.FileName, n,
					errors.New("list item cannot be empty"),
				),
			)
			continue
		}

		r, d := v.listItemVisitor.VisitItem(ctx, opts, n)
		diags = append(diags, d...)

		if !d.HasError() {
			out = append(out, r)
		}
	}

	return out, diags
}

type ObjectVisitor[T any] struct {
	fields       []FieldVisitor[T]
	fieldsByName map[string]struct{}
	constructor  func(context.Context, *T) error
	validator    func(ctx context.Context, n ast.Node, dst *T) error
}

func (v *ObjectVisitor[T]) handleUnknownField(opts *TraverseOpts, key string, node *ast.MappingValueNode) *parsetypes.Diagnostic {
	if opts.UnknownFieldAction == UnknownFieldActionIgnore {
		return nil
	}

	err := parsetypes.NewAnnotatedError(fmt.Errorf("unknown field %q", key), "remove unknown property")
	diag := newErrDiagnosticFromMapping(opts, node, err)
	if opts.UnknownFieldAction == UnknownFieldActionWarn {
		diag.Severity = parsetypes.DiagnosticSeverityWarning
	}

	return diag
}

// Constructor initializes object using given func before mapping.
func (v *ObjectVisitor[T]) Constructor(fn func(context.Context, *T) error) *ObjectVisitor[T] {
	v.constructor = fn
	return v
}

// Validation adds object validation after all fields are mapped.
func (v *ObjectVisitor[T]) Validation(fn func(context.Context, ast.Node, *T) error) *ObjectVisitor[T] {
	v.validator = fn
	return v
}

func (v *ObjectVisitor[T]) VisitItem(ctx context.Context, opts *TraverseOpts, node ast.Node) (T, parsetypes.Diagnostics) {
	nodes := make(map[string]*ast.MappingValueNode, len(v.fields))

	var (
		out   T
		diags parsetypes.Diagnostics
	)

	if v.constructor != nil {
		if err := v.constructor(ctx, &out); err != nil {
			diags = append(diags, NewErrDiagnosticFromNode(opts.FileName, node, err))
		}
	}

	mn, ok := node.(*ast.MappingNode)
	if !ok {
		diags = append(diags,
			NewErrDiagnosticFromNode(
				opts.FileName, node,
				fmt.Errorf("node should be an object, got %s", node.Type()),
			),
		)

		return out, diags
	}

	// Collect field members & validate key names
	for _, n := range mn.Values {
		kn, ok := n.Key.(*ast.StringNode)
		if !ok {
			diags = append(diags,
				newErrDiagnosticFromMapping(opts, n, errors.New("key should be a string")),
			)

			continue
		}

		if _, ok := v.fieldsByName[kn.Value]; !ok {
			if diag := v.handleUnknownField(opts, kn.Value, n); diag != nil {
				diags = append(diags, diag)
			}
			continue
		}

		if _, ok := nodes[kn.Value]; ok {
			diags = append(diags,
				newErrDiagnosticFromMapping(opts, n, fmt.Errorf("duplicate field %q", kn.Value)),
			)
			continue
		}

		nodes[kn.Value] = n
	}

	// Map fields respecting schema order
	for _, dec := range v.fields {
		key := dec.Name()
		n, ok := nodes[key]
		if !ok {
			if dec.IsRequired() {
				diags = append(diags,
					NewErrDiagnosticFromNode(
						opts.FileName, node,
						fmt.Errorf("field %q is required", key),
					),
				)
			}
			continue
		}

		fieldDiags := dec.VisitField(ctx, opts, n, &out)
		diags = append(diags, fieldDiags...)
		if fieldDiags.HasError() {
			continue
		}

		if err := dec.Validate(ctx, &out); err != nil {
			diags = append(diags,
				NewErrDiagnosticFromNode(opts.FileName, n, err),
			)
		}
	}

	if !diags.HasError() && v.validator != nil {
		if err := v.validator(ctx, node, &out); err != nil {
			diags = append(diags, NewErrDiagnosticFromNode(opts.FileName, node, err))
		}
	}

	return out, diags
}
