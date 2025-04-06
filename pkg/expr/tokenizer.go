package expr

import (
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/go-gilbert/gilbert/pkg/iterutil"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type DocumentInfo struct {
	// FileName is document filename to which expression belongs.
	FileName string `json:"fileName"`

	// ByteOffset is Offset in bytes where expression starts.
	//
	// This value affects all Offset numbers returned by Tokenizer and errors.
	ByteOffset int `json:"byteOffset"`

	// StartPosition is line and column number of where expression starts.
	//
	// Added to expression nodes location information.
	StartPosition parsetypes.Position `json:"startPosition"`
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
	lastLinePos  parsetypes.Position
	exprStarted  TokenType
}

// NewTokenizer constructs a new tokenizer.
func NewTokenizer(src string, opts ...Option) *Tokenizer {
	cfg := newParseConfig(opts)
	return newTokenizer(cfg, src)
}

func newTokenizer(cfg parseConfig, src string) *Tokenizer {
	return &Tokenizer{
		docInfo:     cfg.docInfo,
		src:         src,
		lastLinePos: cfg.docInfo.StartPosition,
		stack:       make([]*stackEntry, 0, 10),
	}
}

func (t *Tokenizer) pos() parsetypes.Position {
	if t.prevToken == nil {
		return parsetypes.NewEmptyPosition()
	}

	return t.prevToken.Range.End
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
		tok.Prev = t.prevToken
		t.prevToken = tok.Prev
	}

	t.prevToken = tok
}

// IterTokens returns an iterator to loop through tokens.
func (t *Tokenizer) IterTokens() iter.Seq2[*Token, *TokenError] {
	return func(yield func(*Token, *TokenError) bool) {
		for {
			tok := t.Next()
			if t.err != nil {
				yield(tok, t.err)
				return
			}

			if tok == nil {
				return
			}

			if !yield(tok, t.err) {
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
	if t.err != nil {
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
	lastIndex := len(t.src) - 1
	for i := t.offset; i < len(t.src); i++ {
		//fmt.Printf("%3d; pos:%5s %c\n", i, pos, t.src[i])
		switch char := t.src[i]; char {
		case '\n':
			// TODO: consume windows \r\n as one char.
			endOffset = i
			t.lastLinePos = pos
			pos.Line++
			pos.Column = 1
			t.updateStackEndPos(pos)
		case exprPrefix:
			tok, nextOffset, err := t.checkOpenToken(i, pos)
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

			t.offset = nextOffset
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
				if i != lastIndex {
					pos.Column++
				}

				t.updateStackEndPos(pos)
				continue
			}

			// expression closed
			t.offset = endTok.Offset + len(endTok.Content) - 1 // FIXME: -2?
			if endTok.Prev != nil && endTok.Prev.Type == TokenTypeString {
				// return contents first and then expr terminator.
				t.pendingToken = endTok
				endTok = endTok.Prev
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
	tok := &Token{
		Prev:    t.prevToken,
		Type:    TokenTypeString,
		Content: t.src[t.offset:],
		Offset:  t.offset,
		Range:   parsetypes.NewRange(startPos, pos),
	}

	t.offset = endOffset
	t.setPrevToken(tok)
	return tok
}

// checkCloseToken checks if there are expression close parentheses at given offset.
//
// Passed pos should point to a previous char.
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

	// NOTE: passed pos still points to a prev char.
	closeTokRange := parsetypes.Range{
		Start: pos.Add(0, 1),
		End:   pos.Add(0, len(tokStr)),
	}

	if !openTok.typ.isClosedBy(tokTyp) {
		// Break if close and open tokens don't match
		return nil, &TokenError{
			//Position: tokRange,
			Position: closeTokRange,
			Offset:   t.docInfo.newOffsetRange(offset, offset+len(tokStr)-1),
			Err:      fmt.Errorf("unexpected Token %q", tokStr),
			Note:     noteRemoveToken(tokStr),
		}
	}

	// mark expression closed.
	//exprRange := openTok.rng.WithEndPosition(closeTokRange.End)
	t.popExprStack(closeTokRange.End)
	closeTok := &Token{
		Prev:    t.prevToken,
		Type:    tokTyp,
		Content: tokStr,
		Offset:  offset + t.docInfo.ByteOffset,
		Range:   closeTokRange,
	}

	// Just return a Token if there is no Content inside expression.
	contentLen := closeTok.Offset - t.prevToken.EndOffset() - 1
	if contentLen == 0 {
		return closeTok, nil
	}

	i := offset - contentLen
	str := t.src[i:offset]

	// String might start from a new line.
	startPos := t.prevToken.Range.End
	if str[0] == '\n' {
		// TODO: handle \r\n
		startPos.Line++
		startPos.Column = 1
	} else {
		startPos.Column++
	}

	// Close token can be at line start
	endPos := pos
	//switch {
	//case len(str) == 1:
	//	endPos = startPos
	//case iterutil.LastChar(str) == '\n':
	//	// TODO: handle \r\n
	//	endPos = t.lastLinePos
	//default:
	//	endPos = endPos.Add(0, -1)
	//}

	// make string contents parent of close Token
	strTok := &Token{
		Prev:    t.prevToken,
		Type:    TokenTypeString,
		Content: str,
		Offset:  i + t.docInfo.ByteOffset,
		Range:   parsetypes.NewRange(startPos, endPos),
	}

	closeTok.Prev = strTok
	return closeTok, nil
}

// checkOpenToken checks whether at current offset there an open expression start and returns it as a Token.
//
// Note: pos should be a position of a previous character, not an expression delimiter (`$`)!
func (t *Tokenizer) checkOpenToken(offset int, pos parsetypes.Position) (*Token, int, *TokenError) {
	var (
		tokTyp TokenType
		tokStr string
	)

	switch {
	case hasNextPrefix(t.src, offset+1, evalStartTok):
		tokTyp = TokenTypeEvalStart
		tokStr = evalStartTok
	case hasNextPrefix(t.src, offset+1, shellStartTok):
		tokTyp = TokenTypeShellStart
		tokStr = shellStartTok
	default:
		return nil, 0, nil
	}

	tokContent := t.src[offset : offset+len(tokStr)+1]
	nextOffset := offset + len(tokStr) + 1
	tok := &Token{
		Prev:    t.prevToken,
		Type:    tokTyp,
		Content: tokContent,
		Offset:  offset + t.docInfo.ByteOffset,
		Range: parsetypes.NewRange(
			pos,
			pos.Add(0, len(tokStr)),
		),
	}

	if err := t.validateExprStart(tok); err != nil {
		return nil, 0, err
	}

	t.pushExprStack(offset, tok)
	if offset-t.offset > 1 {
		// consume string left behind
		prevPos := t.pos()
		strContent := t.src[t.offset:offset]

		strEndPos := pos.Sub(0, 1)
		if iterutil.LastChar(strContent) == '\n' {
			strEndPos = t.lastLinePos
		}

		//strContent := t.src[t.offset : offset-1]
		strTok := &Token{
			Prev:    t.prevToken,
			Type:    TokenTypeString,
			Content: strContent,
			Offset:  t.offset + t.docInfo.ByteOffset,
			Range: parsetypes.NewRange(
				prevPos, strEndPos,
			),
		}

		tok.Prev = strTok
		t.pendingToken = tok
		return strTok, nextOffset, nil
	}

	return tok, nextOffset, nil
}

func (t *Tokenizer) validateExprStart(tok *Token) *TokenError {
	openTok := t.lastOpenExpr()
	if openTok == nil {
		return nil
	}

	offset := parsetypes.NewOffsetRange(tok.Offset, len(tok.Content)-1)
	switch openTok.typ {
	case TokenTypeEvalStart:
		return &TokenError{
			Position: tok.Range,
			Offset:   offset,
			Err:      newUnexpectedTokenErr(tok),
			Note:     "eval expression cannot contain another expression",
		}
	case TokenTypeShellStart:
		if tok.Type == openTok.typ {
			return &TokenError{
				Position: tok.Range,
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
	t.updateStackEndPos(tok.Range.End)
	t.stack = append(t.stack, &stackEntry{
		typ:    tok.Type,
		rng:    tok.Range,
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
