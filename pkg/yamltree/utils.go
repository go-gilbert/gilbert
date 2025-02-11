package yamltree

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

func intoDictNode(fi *TraverseOpts, node ast.Node) (*ast.MappingNode, parsetypes.Diagnostics) {
	mn, ok := node.(*ast.MappingNode)
	if !ok {
		return nil, parsetypes.Diagnostics{
			newErrDiagnosticFromNode(
				fi.FileName, node,
				errors.New("node should be a dictionary"),
			),
		}
	}

	return mn, nil
}

// IsNullNode checks whether node type is null.
func IsNullNode(n ast.Node) bool {
	return n.Type() == ast.NullType
}

// IsPrimitiveNode returns whether node is a string, bool or other scalar type.
func IsPrimitiveNode(n ast.Node) bool {
	switch n.Type() {
	case ast.StringType, ast.IntegerType, ast.FloatType, ast.BoolType, ast.LiteralType:
		return true
	default:
		return false
	}
}

// IsStringNode returns whether node is a string (including multiline).
func IsStringNode(n ast.Node) bool {
	switch n.Type() {
	case ast.StringType, ast.LiteralType:
		return true
	default:
		return false
	}
}

// GetNodeRange returns node document range
func GetNodeRange(node ast.Node) (parsetypes.Range, parsetypes.OffsetRange) {
	startTok := node.GetToken()
	startPos := startTok.Position
	rng := parsetypes.NewRange(
		parsetypes.NewPosition(startPos.Line, startPos.Column),
		parsetypes.NewPosition(startPos.Line, startPos.Column),
	)

	offset := parsetypes.OffsetRange{
		Start: startPos.Offset,
		End:   startPos.Offset,
	}

	switch n := node.(type) {
	case *ast.StringNode,
		*ast.IntegerNode,
		*ast.FloatNode,
		*ast.BoolNode:
		strlen := len(startTok.Value)
		rng.End.Column += strlen
		offset.End += strlen
		break
	case *ast.LiteralNode:
		rng, offset = getLiteralRange(n)
	}

	return rng, offset
}

type Source struct {
	// FilePath is file name passed to decoders and provide correct diagnostic messages.
	//
	// If Reader is nil - used to open a source file for read.
	FilePath string

	// Reader is source stream to read.
	//
	// When nil - reader attempts to read a file using provided FilePath.
	Reader io.Reader
}

func (src Source) getReader() (io.Reader, error) {
	if src.Reader != nil {
		return src.Reader, nil
	}

	f, err := os.Open(src.FilePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

type ReadOption = func(*TraverseOpts)

// WithUnknownFieldAction option configures unknown struct fields handling strategy.
func WithUnknownFieldAction(action UnknownFieldAction) ReadOption {
	return func(opts *TraverseOpts) {
		opts.UnknownFieldAction = action
	}
}

// ReadSource unmarshals and decodes YAML from stream.
func ReadSource[T any](ctx context.Context, v ValueVisitor[T], src Source, opts ...ReadOption) (val T, diags parsetypes.Diagnostics, err error) {
	r, err := src.getReader()
	if err != nil {
		return val, diags, err
	}

	if closer, ok := r.(io.Closer); ok {
		defer closer.Close()
	}

	var node ast.Node
	if err := yaml.NewDecoder(r).DecodeContext(ctx, &node); err != nil {
		// TODO: map yaml errors to diagnostics
		return val, diags, err
	}

	cfg := &TraverseOpts{FileName: src.FilePath}
	for _, opt := range opts {
		opt(cfg)
	}

	val, diags = v.VisitItem(ctx, cfg, node)

	return val, diags, nil
}
