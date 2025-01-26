package expr

import (
	"bytes"
	"errors"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/compiler"
	"github.com/expr-lang/expr/conf"
	"github.com/expr-lang/expr/parser"
)

type Range struct {
	StartCol int
	EndCol   int
}

func (r Range) Empty() bool {
	return r.StartCol == r.EndCol && r.EndCol == 0
}

func NewRange(start, end int) Range {
	return Range{
		StartCol: start,
		EndCol:   end,
	}
}

// Expression represents dynamic expression value that can be evaluated in runtime.
type Expression interface {
	// Evaluable returns whether expression is literal or should be evaluated.
	//
	// Can be used to restrict usage of dynamic expressions.
	Evaluable() bool

	// Range returns start and end position of expression.
	Range() Range

	// Eval evaluates an expression and returns a value.
	Eval(ctx EvalContext) (any, error)

	// String returns string representation of evaluated value.
	String(ctx EvalContext) ([]byte, error)
}

type expressionHeader struct {
	Pos Range
}

func newExpressionHeader(pos Range) expressionHeader {
	return expressionHeader{
		Pos: pos,
	}
}

func (h expressionHeader) Range() Range {
	return h.Pos
}

// EmptyExpression represents an empty statement.
type EmptyExpression struct{}

func (e EmptyExpression) Evaluable() bool {
	return false
}

func (e EmptyExpression) Range() Range {
	return Range{}
}

func (e EmptyExpression) String(_ EvalContext) ([]byte, error) {
	return nil, nil
}

func (e EmptyExpression) Eval(_ EvalContext) (any, error) {
	return nil, nil
}

// LiteralExpression represents a literal string.
//
// Example:
//
//	"foobar"
type LiteralExpression struct {
	expressionHeader
	Value string
}

func NewLiteralExpression(pos Range, value string) *LiteralExpression {
	return &LiteralExpression{
		expressionHeader: newExpressionHeader(pos),
		Value:            value,
	}
}

func (l LiteralExpression) Evaluable() bool {
	return false
}

func (l LiteralExpression) Eval(_ EvalContext) (any, error) {
	return l.Value, nil
}

func (l LiteralExpression) String(_ EvalContext) ([]byte, error) {
	return []byte(l.Value), nil
}

// EvalExpression represents evaluation expression.
//
// Example:
//
//	"${foo.bar}"
type EvalExpression struct {
	expressionHeader
	AST        *parser.Tree
	EvalConfig *conf.Config
}

func NewEvalExpression(pos Range, exprStr string, cfg *conf.Config) (*EvalExpression, error) {
	if cfg == nil {
		cfg = evalConfWithOptions()
	}

	tree, err := parseEvalExpr(cfg, exprStr)
	if err != nil {
		return nil, err
	}

	return &EvalExpression{
		expressionHeader: newExpressionHeader(pos),
		AST:              tree,
		EvalConfig:       cfg,
	}, nil
}

func (ee EvalExpression) Evaluable() bool {
	return true
}

func (ee EvalExpression) Eval(ctx EvalContext) (any, error) {
	program, err := compiler.Compile(ee.AST, ee.EvalConfig)
	if err != nil {
		// TODO: unwrap syntax errors
		return nil, newExprError(err, ee.Range())
	}

	vals := ctx.Env.Values()
	output, err := expr.Run(program, vals)
	if err != nil {
		return nil, newExprError(err, ee.Range())
	}

	return output, nil
}

func (ee EvalExpression) String(ctx EvalContext) ([]byte, error) {
	result, err := ee.Eval(ctx)
	if err != nil {
		return nil, err
	}

	v, err := valueToString(result)
	if err != nil {
		err = newExprError(err, ee.Range())
	}

	return []byte(v), err
}

// CompositeExpression represents a sequence of expressions concatenated into a string.
//
// Example:
//
//	"foo ${bar} $(baz)"
type CompositeExpression struct {
	expressionHeader
	Parts []Expression
}

func NewCompositeExpression(rng Range, parts []Expression) *CompositeExpression {
	return &CompositeExpression{
		expressionHeader: newExpressionHeader(rng),
		Parts:            parts,
	}
}

func (ce CompositeExpression) Evaluable() bool {
	return true
}

func (ce CompositeExpression) Eval(ctx EvalContext) (any, error) {
	return ce.String(ctx)
}

func (ce CompositeExpression) String(ctx EvalContext) ([]byte, error) {
	sb := &bytes.Buffer{}
	for _, e := range ce.Parts {
		val, err := e.String(ctx)
		if err != nil {
			return nil, decorateExprError(err, ce.Range())
		}

		sb.Write(val)
	}

	return sb.Bytes(), nil
}

// ShellExpression represents a shell call expression.
// Expression permits nested interpolation expressions.
//
// Example:
//
//	"$(uname -m)"
//	"$(ls ${env.HOME})"
type ShellExpression struct {
	expressionHeader
	Parts []Expression
}

func NewShellExpression(pos Range, parts []Expression) *ShellExpression {
	return &ShellExpression{
		expressionHeader: newExpressionHeader(pos),
		Parts:            parts,
	}
}

func (se ShellExpression) Evaluable() bool {
	return true
}

func (se ShellExpression) Eval(ctx EvalContext) (any, error) {
	r, err := se.String(ctx)
	if err != nil {
		return nil, err
	}

	return string(r), nil
}

func (se ShellExpression) String(ctx EvalContext) ([]byte, error) {
	sb := &strings.Builder{}

	for _, e := range se.Parts {
		val, err := e.String(ctx)
		if err != nil {
			return nil, decorateExprError(err, se.Range())
		}

		sb.Write(val)
	}

	cmd := strings.TrimSpace(sb.String())
	result, err := ctx.CommandProcessor.EvalCommand(cmd)
	if err != nil {
		return nil, newExprError(err, se.Range())
	}

	result = bytes.TrimSpace(result)
	return result, nil
}

// decorateExprError adds parent range to an error it's an expression error.
func decorateExprError(err error, r Range) error {
	var e *ExpressionError
	if !errors.As(err, &e) || r.Empty() {
		return err
	}

	e.ParentRange = r
	return e
}
