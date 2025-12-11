package expr

import (
	"context"
	"errors"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/compiler"
	"github.com/expr-lang/expr/conf"
	"github.com/expr-lang/expr/file"
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

	// EvalText evaluates an expression and returns a string representation as bytes.
	//
	// Returns an error if result cannot be converted to string.
	EvalText(ctx context.Context, p EvalParams) ([]byte, *parsetypes.Diagnostic)
}

type Header struct {
	Source parsetypes.Location
}

func newHeader(loc parsetypes.Location) Header {
	return Header{
		Source: loc,
	}
}

func (h Header) Location() parsetypes.Location {
	return h.Source
}

type EmptyExpression struct {
	Header
}

func NewEmptyExpression(loc parsetypes.Location) *EmptyExpression {
	return &EmptyExpression{
		Header: newHeader(loc),
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
	Header

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
		Header: newHeader(loc),
		Parts:  parts,
	}
}

func (exp *CompositeExpression) Evaluable() bool {
	for _, part := range exp.Parts {
		if part.Evaluable() {
			return true
		}
	}

	return false
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
	Header
	Value string
}

func NewStringExpression(loc parsetypes.Location, value string) *StringExpression {
	return &StringExpression{
		Header: newHeader(loc),
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
	Header

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
		cfg = conf.CreateNew()
	}

	contentLoc := body.Location()
	exp := &EvalExpression{
		Header:          newHeader(loc),
		EvalConfig:      cfg,
		ContentPosition: contentLoc.Range,
		ContentOffset:   contentLoc.Offset,
	}

	var err error
	exp.AST, err = evalExprFromBody(cfg, body)
	if err != nil {
		return nil, exp.errorToDiagnostic(err, "syntax error in eval expression")
	}

	return exp, nil
}

func (exp *EvalExpression) errorToDiagnostic(err error, msg string) *parsetypes.Diagnostic {
	diag := &parsetypes.Diagnostic{
		FileName: exp.Source.FileName,
		Severity: parsetypes.DiagnosticSeverityError,
		Range:    exp.ContentPosition,
		Offset:   exp.ContentOffset,
		Err:      err,
	}

	var t *file.Error
	switch {
	case errors.As(err, &t):
		// TODO: figure out if t.Location is actually useful.
		lineNo := t.Line - 1
		startCol := t.Location.From
		if diag.Range.Start.Column > 0 {
			startCol--
		}

		// set offset only for a first line as line bounds not known.
		if t.Line == 1 {
			startOffset := diag.Offset.Start + t.Location.From - 1
			endOffset := diag.Offset.Start + t.Location.To - 1
			diag.Offset = parsetypes.NewOffsetRange(startOffset, endOffset)
		}

		startPos := diag.Range.Start.Add(lineNo, startCol)
		endPos := startPos.Add(0, t.Location.To-t.Location.From)
		diag.Range = parsetypes.NewRange(startPos, endPos)
		diag.Note = t.Message
		diag.Err = fmt.Errorf("%s: %s", msg, t.Message)
	}

	return diag
}

func (exp *EvalExpression) Eval(_ context.Context, p EvalParams) (any, *parsetypes.Diagnostic) {
	prog, err := compiler.Compile(exp.AST, exp.EvalConfig)
	if err != nil {
		// FIXME: map error to diagnostics
		return nil, &parsetypes.Diagnostic{
			FileName: exp.Source.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    exp.ContentPosition,
			Offset:   exp.ContentOffset,
			Err:      err,
		}
	}

	vals, err := p.Env.Values()
	if err != nil {
		return nil, &parsetypes.Diagnostic{
			FileName: exp.Source.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    exp.ContentPosition,
			Offset:   exp.ContentOffset,
			Err:      fmt.Errorf("cannot expand scope value: %w", err),
		}
	}

	output, err := expr.Run(prog, vals)
	if err != nil {
		return nil, exp.errorToDiagnostic(err, "expression error")
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
			FileName: exp.Source.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    exp.ContentPosition,
			Offset:   exp.ContentOffset,
			Err:      err,
		}
	}

	return b, nil
}

type ShellExpression struct {
	Header

	Body Expression
}

func NewShellExpression(loc parsetypes.Location, body Expression) *ShellExpression {
	return &ShellExpression{
		Header: newHeader(loc),
		Body:   body,
	}
}

func (exp *ShellExpression) Evaluable() bool {
	return exp.Body != nil
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
		loc := exp.Body.Location()
		return nil, &parsetypes.Diagnostic{
			FileName: exp.Source.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    loc.Range,
			Offset:   loc.Offset,
			Note:     parsetypes.NoteFromError(err),
			Err:      fmt.Errorf("shell expression error: %w", err),
		}
	}

	return r, nil
}
