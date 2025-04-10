package expr

import (
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

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
	for i := t.offset; i < len(t.src); i++ {
		switch char := t.src[i]; char {
		case '\r', '\n':
			tok, nextOffset, err := t.consumeEOL(i, pos)
			if err != nil {
				t.err = err
				return nil
			}

			t.updateStackEndPos(pos)
			t.setPrevToken(tok)
			t.offset = nextOffset
			return tok

		case exprPrefix:
			pos.Column++
			endOffset = i
			tok, newOffset, err := t.checkOpenToken(i, pos)
			if err != nil {
				t.err = err
				return nil
			}

			if tok == nil {
				t.updateStackEndPos(pos)
				continue
			}

			t.offset = newOffset
			t.setPrevToken(tok)
			return tok

		default:
			pos.Column++
			endOffset = i
			endTok, nextOffset, err := t.checkCloseToken(i, pos)
			if err != nil {
				t.err = err
				return nil
			}

			if endTok == nil {
				t.updateStackEndPos(pos)
				continue
			}

			// If string was before close token:
			t.offset = nextOffset
			if endTok.Prev != nil && endTok.Prev.Type == TokenTypeString {
				t.pendingToken = endTok
				endTok = endTok.Prev
			}

			t.setPrevToken(endTok)
			return endTok
		}
	}

	openTok := t.lastOpenExpr()
	if openTok != nil {
		// Missing token to close last expression clause.
		t.err = &TokenError{
			Position: openTok.rng,
			Offset:   t.docInfo.newOffsetRange(openTok.offset, endOffset),
			Err:      errors.New("unexpected end of input"),
			Note:     fmt.Sprintf("missing %q", openTok.typ.GetPair().Text()),
		}
		return nil
	}

	if t.offset == endOffset {
		return nil
	}

	// assembly string left behind (if any left)
	strChunk := t.src[t.offset:]
	if strChunk == "" {
		return nil
	}

	startPos = startPos.Add(0, 1)
	tok := &Token{
		Prev:    t.prevToken,
		Type:    TokenTypeString,
		Content: t.src[t.offset:],
		Offset:  t.offset + t.docInfo.ByteOffset,
		Range:   parsetypes.NewRange(startPos, pos),
	}

	t.offset = endOffset
	t.setPrevToken(tok)
	return tok
}

// checkCloseToken checks if there are expression close parentheses at given offset.
//
// Returns an error if there are parentheses at unexpected place.
// Returns nil if there are no close tokens found.
func (t *Tokenizer) checkCloseToken(offset int, pos parsetypes.Position) (*Token, int, *TokenError) {
	// TODO: support escapes?
	tokTyp, tokStr := hasTokenClosePrefix(t.src[offset:])
	if tokTyp == TokenTypeEmpty {
		return nil, 0, nil
	}

	openTok := t.lastOpenExpr()
	if t.prevToken == nil || openTok == nil {
		// outside of expression, we don't care
		return nil, 0, nil
	}

	closeTokRange := parsetypes.Range{
		Start: pos,
		End:   pos.Add(0, len(tokStr)-1),
	}

	if !openTok.typ.IsClosedBy(tokTyp) {
		// Break if close and open tokens don't match
		return nil, 0, &TokenError{
			//Position: tokRange,
			Position: closeTokRange,
			Offset:   t.docInfo.newOffsetRange(offset, offset+len(tokStr)-1),
			Err:      fmt.Errorf("unexpected token %q", tokStr),
			Note:     fmt.Sprintf("expected %q", openTok.typ.GetPair().Text()),
		}
	}

	// mark expression closed.
	t.popExprStack(closeTokRange.End)
	closeTok := &Token{
		Prev:    t.prevToken,
		Type:    tokTyp,
		Content: tokStr,
		Offset:  offset + t.docInfo.ByteOffset,
		Range:   closeTokRange,
	}

	// Just return a Token if there is no Content inside expression.
	nextOffset := offset + len(tokStr)
	strLen := closeTok.Offset - t.prevToken.EndOffset() - 1

	if strLen < 0 {
		// bug assertion
		panic(fmt.Sprintf("checkCloseToken: strLen is less than zero (got: %d)", strLen))
	}

	if strLen == 0 {
		return closeTok, nextOffset, nil
	}

	// Read string behind
	i := offset - strLen
	str := t.src[i:offset]

	strStartPos := t.prevToken.Range.End.Add(0, 1)
	strEndPos := strStartPos.Add(0, len(str)-1)

	// make string contents parent of close Token
	strTok := &Token{
		Prev:    t.prevToken,
		Type:    TokenTypeString,
		Content: str,
		Offset:  i + t.docInfo.ByteOffset,
		Range:   parsetypes.NewRange(strStartPos, strEndPos),
	}

	closeTok.Prev = strTok
	return closeTok, nextOffset, nil
}

// checkOpenToken checks whether at current offset there an open expression start and returns it as a Token.
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

	nextOffset := offset + len(tokStr) + 1
	tokContent := t.src[offset:nextOffset]
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
	if offset-t.offset == 0 {
		return tok, nextOffset, nil
	}

	// consume string left behind
	prevPos := t.pos()
	strContent := t.src[t.offset:offset]
	strStartPos := prevPos.Add(0, 1)
	strEndPos := pos.Sub(0, 1)

	strTok := &Token{
		Prev:    t.prevToken,
		Type:    TokenTypeString,
		Content: strContent,
		Offset:  t.offset + t.docInfo.ByteOffset,
		Range: parsetypes.NewRange(
			strStartPos, strEndPos,
		),
	}

	tok.Prev = strTok
	t.pendingToken = tok
	return strTok, nextOffset, nil
}

func (t *Tokenizer) validateExprStart(tok *Token) *TokenError {
	openTok := t.lastOpenExpr()
	if openTok == nil {
		return nil
	}

	offset := parsetypes.NewOffsetRangeFromLen(tok.Offset, len(tok.Content)-1)
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

func (t *Tokenizer) consumeEOL(offset int, curPos parsetypes.Position) (*Token, int, *TokenError) {
	lineCount := 0
	j := offset
	crSet := false
loop:
	for i := offset; i < len(t.src); i++ {
		switch char := t.src[i]; char {
		case '\r':
			// just for windows.
			j = i
			if crSet {
				// double \r\r means broken newline.
				break loop
			}

			crSet = true
			continue
		case '\n':
			j = i
			crSet = false
			lineCount++
		default:
			break loop
		}
	}

	nextOffset := j + 1
	content := t.src[offset:nextOffset]
	startPos := parsetypes.NewPosition(curPos.Line+1, 0)
	endPos := startPos.Add(lineCount-1, 0)
	if lineCount <= 1 {
		endPos = startPos
	}

	// if for some reason there is unexpected value between CRLF newline:
	if crSet {
		return nil, 0, &TokenError{
			Err: errors.New(`broken CRLF - expected "\n" after "\r"`),
			Position: parsetypes.Range{
				Start: startPos,
				End:   endPos,
			},
			Offset: t.docInfo.newOffsetRange(
				offset, j,
			),
		}
	}

	eolTol := &Token{
		Prev:    t.prevToken,
		Type:    TokenTypeEOL,
		Content: content,
		Offset:  t.docInfo.ByteOffset + offset,
		Range: parsetypes.Range{
			Start: startPos,
			End:   endPos,
		},
	}

	// is there any string left behind?
	str := t.src[t.offset:offset] // offset points to first \n
	if str == "" {
		return eolTol, nextOffset, nil
	}

	strEndPos := curPos // curPos points to a char before newline
	strStartPos := curPos.Sub(0, len(str)-1)
	strTok := &Token{
		Prev:    t.prevToken,
		Type:    TokenTypeString,
		Content: str,
		Offset:  t.docInfo.ByteOffset + t.offset,
		Range: parsetypes.NewRange(
			strStartPos, strEndPos,
		),
	}

	eolTol.Prev = strTok
	t.pendingToken = eolTol
	t.offset = offset

	return strTok, nextOffset, nil
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
