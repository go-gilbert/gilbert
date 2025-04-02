package expr2

import (
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type DocumentInfo struct {
	// FileName is document filename to which expression belongs.
	FileName string

	// ByteOffset is Offset in bytes where expression starts.
	//
	// This value affects all Offset numbers returned by Tokenizer and errors.
	ByteOffset int

	// StartPosition is line and column number of where expression starts.
	//
	// Added to expression nodes location information.
	StartPosition parsetypes.Position
}

func (docInfo DocumentInfo) newOffsetRange(start, end int) parsetypes.OffsetRange {
	return parsetypes.OffsetRange{
		Start: docInfo.ByteOffset + start,
		End:   docInfo.ByteOffset + end,
	}
}

func (docInfo DocumentInfo) translateRange(rng parsetypes.OffsetRange) parsetypes.OffsetRange {
	return parsetypes.OffsetRange{
		Start: docInfo.ByteOffset + rng.Start,
		End:   docInfo.ByteOffset + rng.End,
	}
}

type Option = func(cfg *parseConfig)

type parseConfig struct {
	docInfo DocumentInfo
}

func newParseConfig(opts []Option) parseConfig {
	cfg := parseConfig{
		docInfo: DocumentInfo{
			FileName:      "",
			ByteOffset:    0,
			StartPosition: parsetypes.NewEmptyPosition(),
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

type stackEntry struct {
	typ    TokenType
	rng    parsetypes.Range
	offset int
}

type Tokenizer struct {
	docInfo      DocumentInfo
	stack        []*stackEntry
	prevToken    *Token
	pendingToken *Token
	src          string
	offset       int
	err          *TokenError
}

// NewTokenizer constructs a new tokenizer.
func NewTokenizer(src string, opts ...Option) *Tokenizer {
	cfg := newParseConfig(opts)
	return newTokenizer(cfg, src)
}

func newTokenizer(cfg parseConfig, src string) *Tokenizer {
	return &Tokenizer{
		docInfo: cfg.docInfo,
		src:     src,
		stack:   make([]*stackEntry, 0, 10),
	}
}

func (t *Tokenizer) pos() parsetypes.Position {
	if t.prevToken == nil {
		return parsetypes.NewEmptyPosition()
	}

	return t.prevToken.rng.End
}

func (t *Tokenizer) lastOpenExpr() *stackEntry {
	if len(t.stack) == 0 {
		return nil
	}

	return t.stack[len(t.stack)-1]
}

func (t *Tokenizer) setPrevToken(tok *Token) {
	if tok == nil {
		panic("setPrevToken: nil Token")
	}

	if t.prevToken != nil {
		t.prevToken.prev = tok
	}

	t.prevToken = tok
}

// IterTokens returns an iterator to loop through tokens.
func (t *Tokenizer) IterTokens() iter.Seq2[*Token, *TokenError] {
	return func(yield func(*Token, *TokenError) bool) {
		for {
			tok := t.Next()
			if !yield(tok, t.err) {
				return
			}

			if tok == nil {
				return
			}
		}
	}
}

// Err returns last tokenizer error
func (t *Tokenizer) Err() *TokenError {
	return t.err
}

// Next searches and returns a next token.
func (t *Tokenizer) Next() *Token {
	if t.err == nil {
		return nil
	}

	if t.pendingToken != nil {
		// return Token placed after adjacent string
		t.setPrevToken(t.pendingToken)
		tok := t.pendingToken
		t.pendingToken = nil
		return tok
	}

	startPos := t.pos()
	pos := startPos

	endOffset := t.offset
	//strStartOffset := -1
	for i := t.offset; i < len(t.src); i++ {
		switch char := t.src[i]; char {
		case '\n':
			endOffset = i
			pos.Line++
			pos.Column = 1
			t.updateStackEndPos(pos)
		case exprPrefix:
			tok, err := t.checkOpenToken(i, pos)
			if err != nil {
				t.err = err
				return nil
			}

			if tok == nil {
				endOffset = i
				pos.Column++
				t.updateStackEndPos(pos)
				continue
			}

			t.offset += len(tok.content) - 1
			t.setPrevToken(tok)
			return tok

		default:
			// is this a close parentheses?
			endTok, err := t.checkCloseToken(i, pos)
			if err != nil {
				t.err = err
				return nil
			}

			if endTok == nil {
				// just a char, skip
				endOffset = i
				t.updateStackEndPos(pos)
				pos.Column++
				continue
			}

			// expression closed
			t.offset = endTok.offset + len(endTok.content) - 1 // FIXME: -2?
			if endTok.prev != nil && endTok.prev.typ == TokenTypeString {
				// return contents first and then expr terminator.
				t.pendingToken = endTok
				endTok = endTok.prev
			}

			t.setPrevToken(endTok)
			return endTok
		}
	}

	if t.offset == endOffset {
		return nil
	}

	openTok := t.lastOpenExpr()
	if openTok != nil {
		// Missing token to close last expression clause.
		t.err = &TokenError{
			Position: openTok.rng,
			Offset:   t.docInfo.newOffsetRange(openTok.offset, endOffset),
			Err:      errors.New("unexpected end of input"),
			Note:     fmt.Sprintf("missing %q", tokenToString(openTok.typ)),
		}
		return nil
	}

	// assembly string left behind
	return &Token{
		prev:    t.prevToken,
		typ:     TokenTypeString,
		content: t.src[t.offset:],
		offset:  t.offset,
		rng:     parsetypes.NewRange(startPos, pos),
	}
}

// checkCloseToken checks if there are expression close parentheses at given offset.
//
// Returns an error if there are parentheses at unexpected place.
// Returns nil if there are no close tokens found.
func (t *Tokenizer) checkCloseToken(offset int, pos parsetypes.Position) (*Token, *TokenError) {
	// TODO: support escapes?
	tokTyp, tokStr := hasTokenClosePrefix(t.src[offset:])
	if tokTyp == TokenTypeEmpty {
		return nil, nil
	}

	openTok := t.lastOpenExpr()
	if t.prevToken == nil || openTok == nil {
		// outside of expression, we don't care
		return nil, nil
	}

	tokRange := parsetypes.NewRange(pos, pos.Add(0, len(tokStr)-1)) // FIXME: -2?
	if !openTok.typ.isClosedBy(tokTyp) {
		// Break if close and open tokens don't match
		return nil, &TokenError{
			Position: tokRange,
			Offset:   t.docInfo.newOffsetRange(offset, offset+len(tokStr)-1),
			Err:      fmt.Errorf("unexpected Token %q", tokStr),
			Note:     noteRemoveToken(tokStr),
		}
	}

	// mark expression closed
	t.popExprStack(tokRange.End)
	closeTok := &Token{
		prev:    t.prevToken,
		typ:     tokTyp,
		content: tokStr,
		offset:  offset,
		rng:     tokRange,
	}

	// Just return a Token if there is no content inside expression.
	if t.prevToken.typ.isClosedBy(tokTyp) {
		return closeTok, nil
	}

	// offsets are prefixed with global Offset
	i := t.prevToken.offset - t.docInfo.ByteOffset + len(t.prevToken.content)

	// make string contents parent of close Token
	closeTok.prev = &Token{
		prev:    t.prevToken,
		typ:     TokenTypeString,
		content: t.src[i:offset],
		offset:  i + t.docInfo.ByteOffset,
		rng:     parsetypes.NewRange(t.prevToken.rng.End, pos),
	}

	return closeTok, nil
}

// checkOpenToken checks whether at current offset there an open expression start and returns it as a Token.
//
// Note: pos should be a position of a previous character, not an expression delimiter (`$`)!
func (t *Tokenizer) checkOpenToken(offset int, pos parsetypes.Position) (*Token, *TokenError) {
	var (
		tokTyp TokenType
		tokStr string
	)

	switch {
	case hasNextPrefix(t.src, offset+1, evalStartTok):
		tokTyp = TokenTypeExprStart
		tokStr = evalStartTok
	case hasNextPrefix(t.src, offset+1, shellStartTok):
		tokTyp = TokenTypeShellStart
		tokStr = shellStartTok
	default:
		return nil, nil
	}

	tok := &Token{
		prev:    t.prevToken,
		typ:     tokTyp,
		content: string(exprPrefix) + tokStr,
		offset:  offset + t.docInfo.ByteOffset,
		rng: parsetypes.NewRange(
			// passed Position references char before expression start.
			pos.Add(0, 1),
			pos.Add(0, len(tokStr)+1),
		),
	}

	if err := t.validateExprStart(tok); err != nil {
		return nil, err
	}

	t.pushExprStack(offset, tok)
	if offset-t.offset > 1 {
		// consume string left behind
		prevPos := t.pos()
		strTok := &Token{
			prev:    t.prevToken,
			typ:     TokenTypeString,
			content: t.src[t.offset : offset-1],
			offset:  t.offset + t.docInfo.ByteOffset,
			rng: parsetypes.NewRange(
				prevPos, pos,
			),
		}

		tok.prev = strTok
		t.pendingToken = tok
		return strTok, nil
	}

	return tok, nil
}

func (t *Tokenizer) validateExprStart(tok *Token) *TokenError {
	openTok := t.lastOpenExpr()
	if openTok == nil {
		return nil
	}

	offset := parsetypes.NewOffsetRange(tok.offset, len(tok.content)-1)
	switch openTok.typ {
	case TokenTypeExprStart:
		return &TokenError{
			Position: tok.rng,
			Offset:   offset,
			Err:      newUnexpectedTokenErr(tok),
			Note:     "eval expression cannot contain another expression",
		}
	case TokenTypeShellStart:
		if tok.typ == openTok.typ {
			return &TokenError{
				Position: tok.rng,
				Offset:   offset,
				Err:      newUnexpectedTokenErr(tok),
				Note:     "shell expression cannot contain another shell expression",
			}
		}
	default:
		break
	}

	return nil
}

// pushExprStack adds a Token to a queue to mark expression open start.
func (t *Tokenizer) pushExprStack(offset int, tok *Token) {
	t.updateStackEndPos(tok.rng.End)
	t.stack = append(t.stack, &stackEntry{
		typ:    tok.typ,
		rng:    tok.rng,
		offset: offset,
	})
}

// updateStackEndPos updates end position of nested expressions stack.
// used to provide errors range when unterminated expression found.
func (t *Tokenizer) updateStackEndPos(pos parsetypes.Position) {
	if len(t.stack) == 0 {
		return
	}

	for _, e := range t.stack {
		e.rng.End = pos
	}
}

// popExprStack removes last open clause element from expressions stack.
func (t *Tokenizer) popExprStack(pos parsetypes.Position) {
	if len(t.stack) == 0 {
		panic("popExprStack: expression stack is empty!")
	}

	// update bounds of parents
	t.stack = t.stack[:len(t.stack)-1]
	t.updateStackEndPos(pos)
}

func hasNextPrefix(str string, start int, pfx string) bool {
	if start < len(str) {
		return strings.HasPrefix(str[start:], pfx)
	}

	return false
}
