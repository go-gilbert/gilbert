package expr

import (
	"context"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/compiler"
	"github.com/expr-lang/expr/conf"
	"github.com/expr-lang/expr/parser"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

// Expression represents dynamic expression value that can be evaluated in runtime.
type Expression interface {
	// Evaluable returns whether expression is literal or should be evaluated.
	//
	// Can be used to restrict usage of dynamic expressions.
	Evaluable() bool

	// Location position and offset where expression is defined.
	Location() parsetypes.Location

	// Eval evaluates an expression and returns a value.
	Eval(ctx context.Context, p EvalParams) (any, *parsetypes.Diagnostic)

	// EvalText returns string bytes representation of evaluated value.
	EvalText(ctx context.Context, p EvalParams) ([]byte, *parsetypes.Diagnostic)
}

type header struct {
	location parsetypes.Location
}

func newHeader(loc parsetypes.Location) header {
	return header{
		location: loc,
	}
}

func (h header) Location() parsetypes.Location {
	return h.location
}

type EmptyExpression struct {
	header
}

func NewEmptyExpression(loc parsetypes.Location) *EmptyExpression {
	return &EmptyExpression{
		header: newHeader(loc),
	}
}

func (*EmptyExpression) Evaluable() bool {
	return false
}

func (*EmptyExpression) Eval(_ context.Context, _ EvalParams) (any, *parsetypes.Diagnostic) {
	return nil, nil
}

func (*EmptyExpression) EvalText(_ context.Context, _ EvalParams) ([]byte, *parsetypes.Diagnostic) {
	return nil, nil
}

// CompositeExpression represents a sequence of expressions concatenated into a string.
//
// Example:
//
//	"foo ${bar} $(baz)"
type CompositeExpression struct {
	header

	Parts []Expression
}

func NewCompositeExpression(parentLoc parsetypes.Location, parts []Expression) *CompositeExpression {
	// Use children bounds if possible. Otherwise, use parent as range.
	loc := parentLoc
	if len(parts) > 0 {
		startLoc := parts[0].Location()
		endLoc := parts[len(parts)-1].Location()

		loc.Range = parsetypes.NewRange(startLoc.Range.Start, endLoc.Range.End)
		loc.Offset = parsetypes.OffsetRange{
			Start: startLoc.Offset.Start,
			End:   endLoc.Offset.End,
		}
	}

	return &CompositeExpression{
		header: newHeader(loc),
		Parts:  parts,
	}
}

func (exp *CompositeExpression) Evaluable() bool {
	for _, part := range exp.Parts {
		if part.Evaluable() {
			return true
		}
	}

	return true
}

func (exp *CompositeExpression) Eval(ctx context.Context, p EvalParams) (any, *parsetypes.Diagnostic) {
	switch len(exp.Parts) {
	case 0:
		return nil, nil
	case 1:
		return exp.Parts[0].Eval(ctx, p)
	}

	result, err := exp.evalParts(ctx, p)
	return bytesAsString(result), err
}

func (exp *CompositeExpression) evalParts(ctx context.Context, p EvalParams) ([]byte, *parsetypes.Diagnostic) {
	res := make([]byte, 0, 128)
	for _, part := range exp.Parts {
		result, err := part.EvalText(ctx, p)
		if err != nil {
			return nil, err
		}

		res = append(res, result...)
	}

	return res, nil
}

func (exp *CompositeExpression) EvalText(ctx context.Context, p EvalParams) ([]byte, *parsetypes.Diagnostic) {
	switch len(exp.Parts) {
	case 0:
		return nil, nil
	case 1:
		return exp.Parts[0].EvalText(ctx, p)
	}

	return exp.evalParts(ctx, p)
}

// StringExpression represents a literal string.
//
// Example:
//
//	"foobar"
type StringExpression struct {
	header
	Value string
}

func NewStringExpression(loc parsetypes.Location, value string) *StringExpression {
	return &StringExpression{
		header: newHeader(loc),
		Value:  value,
	}
}

func (*StringExpression) Evaluable() bool {
	return false
}

func (exp *StringExpression) Eval(_ context.Context, _ EvalParams) (any, *parsetypes.Diagnostic) {
	return exp.Value, nil
}

func (exp *StringExpression) EvalText(_ context.Context, _ EvalParams) ([]byte, *parsetypes.Diagnostic) {
	return []byte(exp.Value), nil
}

// EvalExpression represents evaluation expression.
//
// Example:
//
//	"${{foo.bar}}"
type EvalExpression struct {
	header

	AST             *parser.Tree
	EvalConfig      *conf.Config
	ContentPosition parsetypes.Range
	ContentOffset   parsetypes.OffsetRange
}

func (*EvalExpression) Evaluable() bool {
	return true
}

func NewEvalExpression(loc parsetypes.Location, body Expression, cfg *conf.Config) (*EvalExpression, *parsetypes.Diagnostic) {
	if cfg == nil {
		cfg = evalConfWithOptions()
	}

	contentLoc := body.Location()
	tree, err := evalExprFromBody(cfg, body)
	if err != nil {
		// FIXME: map error to diagnostic
		return nil, &parsetypes.Diagnostic{
			FileName: loc.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    contentLoc.Range,
			Offset:   contentLoc.Offset,
			Err:      err,
		}
	}

	return &EvalExpression{
		header:          newHeader(loc),
		AST:             tree,
		EvalConfig:      cfg,
		ContentPosition: contentLoc.Range,
		ContentOffset:   contentLoc.Offset,
	}, nil
}

func (exp *EvalExpression) Eval(_ context.Context, p EvalParams) (any, *parsetypes.Diagnostic) {
	prog, err := compiler.Compile(exp.AST, exp.EvalConfig)
	if err != nil {
		// FIXME: map error to diagnostics
		return nil, &parsetypes.Diagnostic{
			FileName: exp.location.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    exp.ContentPosition,
			Offset:   exp.ContentOffset,
			Err:      err,
		}
	}

	vals, err := p.Env.Values()
	if err != nil {
		return nil, &parsetypes.Diagnostic{
			FileName: exp.location.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    exp.ContentPosition,
			Offset:   exp.ContentOffset,
			Err:      fmt.Errorf("cannot expand scope value: %w", err),
		}
	}

	output, err := expr.Run(prog, vals)
	if err != nil {
		// FIXME: map error to diagnostics
		return nil, &parsetypes.Diagnostic{
			FileName: exp.location.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    exp.ContentPosition,
			Offset:   exp.ContentOffset,
			Err:      err,
		}
	}

	return output, nil
}

func (exp *EvalExpression) EvalText(ctx context.Context, p EvalParams) ([]byte, *parsetypes.Diagnostic) {
	res, diag := exp.Eval(ctx, p)
	if diag != nil {
		return nil, diag
	}

	b, err := valueToBytes(res)
	if err != nil {
		return nil, &parsetypes.Diagnostic{
			FileName: exp.location.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    exp.ContentPosition,
			Offset:   exp.ContentOffset,
			Err:      err,
		}
	}

	return b, nil
}

type ShellExpression struct {
	header

	Body Expression
}

func NewShellExpression(loc parsetypes.Location, body Expression) *ShellExpression {
	return &ShellExpression{
		header: newHeader(loc),
		Body:   body,
	}
}

func (exp *ShellExpression) Eval(ctx context.Context, p EvalParams) (any, *parsetypes.Diagnostic) {
	result, err := exp.EvalText(ctx, p)
	if err != nil {
		return nil, err
	}

	return bytesAsString(result), nil
}

func (exp *ShellExpression) EvalText(ctx context.Context, p EvalParams) ([]byte, *parsetypes.Diagnostic) {
	if exp.Body == nil {
		return nil, nil
	}

	src, diag := exp.Body.EvalText(ctx, p)
	if diag != nil {
		return nil, diag
	}

	if len(src) == 0 {
		return nil, nil
	}

	r, err := p.CommandProcessor.EvalCommand(ctx, bytesAsString(src))
	if err != nil {
		return nil, &parsetypes.Diagnostic{
			FileName: exp.location.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    exp.header.location.Range,
			Offset:   exp.header.location.Offset,
			Err:      err,
		}
	}

	return r, nil
}

func (exp *ShellExpression) Evaluable() bool {
	return exp.Body != nil
}
