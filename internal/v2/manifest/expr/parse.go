package expr

import (
	"strings"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type token int

const (
	tokenEmpty token = iota
	tokenExprStart
	tokenShellStart
	tokenEnd
)

type documentPos struct {
	fileName  string
	docOffset int
	offset    parsetypes.OffsetRange
	startPos  parsetypes.Position
	endPos    parsetypes.Position
}

func (docPos *documentPos) advanceStartLine() {
	docPos.startPos.Line++
	docPos.startPos.Column = 1
}

func (docPos *documentPos) advanceEndLine() {
	docPos.endPos.Line++
	docPos.endPos.Column = 1
}

func (docPos *documentPos) location() parsetypes.Location {
	return parsetypes.Location{
		FileName: docPos.fileName,
		Range:    parsetypes.NewRange(docPos.startPos, docPos.endPos),
		Offset: parsetypes.OffsetRange{
			Start: docPos.docOffset + docPos.offset.Start,
			End:   docPos.docOffset + docPos.offset.End,
		},
	}
}

type tokenPos struct {
	token     token
	pos       documentPos
	beforePos parsetypes.Range
}

func findOpenToken(curPos documentPos, str string, offset int, stopToken byte) tokenPos {
	// TODO: support escapes?
	n := len(str)

	tokenStartPos := -1
	beforePos := curPos
	for i := offset; i < n; i++ {
		curPos.endPos.Column++

		switch v := str[i]; v {
		case '$':
			tokenStartPos = i
		case '{':
			if tokenStartPos != -1 {
				curPos.startPos = curPos.endPos
				curPos.offset = parsetypes.OffsetRange{
					Start: tokenStartPos,
					End:   i,
				}
				return tokenPos{
					token: tokenExprStart,
					pos:   curPos,
				}
			}
		case '(':
			if tokenStartPos != -1 {
				curPos.startPos = curPos.endPos
				curPos.offset = parsetypes.OffsetRange{
					Start: tokenStartPos,
					End:   i,
				}
				return tokenPos{
					token: tokenShellStart,
					pos:   curPos,
				}
			}
		default:
			tokenStartPos = -1
			if v == '\n' {
				curPos.endPos.Column = 1
				curPos.endPos.Line++
			}

			if stopToken != 0 && v == stopToken {
				curPos.startPos = curPos.endPos
				curPos.offset = parsetypes.OffsetRange{
					Start: i,
					End:   i,
				}
				return tokenPos{
					token: tokenEnd,
					pos:   curPos,
				}
			}
		}
	}

	curPos.offset.Start = offset
	curPos.offset.End = len(str) - 1
	return tokenPos{
		token: tokenEmpty,
		pos:   curPos,
	}
}

// Parse parses string interpolation expression.
//
// Allowed inputs examples:
//
//	"foobar"
//	"foo ${bar.baz}"
//	"2+2=${2+2}!"
//	"OS is $(uname -s)"
func Parse(str string, opts ...Option) (Expression, error) {
	if str == "" {
		return EmptyExpression{}, nil
	}

	cfg := configFromOptions(opts)
	docPos := documentPos{
		fileName:  cfg.docPos.FileName,
		docOffset: cfg.docPos.ByteOffset,
		startPos:  cfg.docPos.Position,
		offset: parsetypes.OffsetRange{
			Start: 0,
			End:   len(str) - 1,
		},
	}

	root := NewCompositeExpression(docPos, make([]Expression, 0, 1))
	start := 0
	end := len(str)
	for start < end {
		tok := findOpenToken(docPos, str, start, 0)
		if tok.token == tokenEmpty {
			// Break if nothing to parse
			strChunk := str[start : tok.pos.offset.End+1]
			root.Parts = append(
				root.Parts,
				NewLiteralExpression(tok.pos, strChunk),
			)
			break
		}

		// Append adjacent string literal if any.
		if tok.pos.offset.Start-1 > start {
			strChunk := str[start:tok.pos.offset.Start]
			strChunkPos := docPos
			// TODO: check when prev tok ends.
			strChunkPos.endPos = tok.pos.endPos
			strChunkPos.offset = parsetypes.OffsetRange{
				Start: start,
				End:   tok.pos.offset.Start - 1,
			}
			root.Parts = append(
				root.Parts,
				NewLiteralExpression(strChunkPos, strChunk),
			)
		}

		endPos, expr, err := consumeToken(docPos, str, tok)
		if err != nil {
			return nil, err
		}

		root.Parts = append(root.Parts, expr)
		docPos = endPos
		docPos.endPos.Column++
		docPos.startPos = docPos.endPos
		docPos.offset = parsetypes.OffsetRange{
			Start: docPos.offset.End + 1,
			End:   docPos.offset.End + 1,
		}
		start = endPos.offset.End
	}

	// Unwrap if there is only one statement
	if len(root.Parts) == 1 {
		return root.Parts[0], nil
	}

	return root, nil
}

func consumeToken(parent documentPos, str string, pos tokenPos) (documentPos, Expression, *ExpressionError) {
	switch pos.token {
	case tokenExprStart:
		return consumeExprToken(parent, str, pos)
	case tokenShellStart:
		return consumeShellToken(parent, str, pos)
	default:
		return parent, nil, newNestedExprError(ErrBadToken, NewRange(pos.rng.Start, pos.rng.End), parent)
	}
}

func consumeExprToken(parent documentPos, str string, tok tokenPos) (documentPos, Expression, *ExpressionError) {
	endPos := -1
	currentPos := tok.pos
	prevEolPos := currentPos.endPos
	for i := tok.pos.offset.End + 1; i < len(str); i++ {
		if str[i] == '\n' {
			prevEolPos = currentPos.endPos
			currentPos.advanceEndLine()
			continue
		}

		currentPos.endPos.Column++
		//prevEolPos.Column++
		if str[i] == '}' {
			endPos = i
			break
		}
	}

	if endPos == -1 {
		currentPos.offset.End = parent.offset.End
		return currentPos, nil, newNestedExprError(
			ErrUnterminatedExpression,
			currentPos,
			parent,
		)
	}

	currentPos.offset.End = endPos
	content := str[tok.pos.offset.End+1 : endPos]
	if content == "" {
		return currentPos, nil, newNestedExprError(
			ErrEmptyExpression,
			currentPos,
			parent,
		)
	}

	// If content is surrounded by new line - trim content position
	contentPos := currentPos
	if content[0] == '\n' {
		contentPos.offset.Start++
		contentPos.advanceStartLine()
		content = content[1:]
	}

	if content[len(content)-1] == '\n' {
		contentPos.offset.End--
		contentPos.endPos = prevEolPos
		content = content[:len(content)-1]
	}

	exp, err := NewEvalExpression(contentPos, content, nil)
	if err != nil {
		return currentPos, nil, newNestedExprError(err, currentPos, parent)
	}

	return currentPos, exp, nil
}

func consumeShellToken(parent documentPos, str string, startTok tokenPos) (documentPos, Expression, *ExpressionError) {
	currentPos := startTok.pos
	currentPos.offset.End++
	currentPos.startPos.Column++

	children := make([]Expression, 0, 1)

	//se := NewShellExpression(
	//	NewRange(startTok.rng.Start, parent.EndCol),
	//	make([]Expression, 0, 1),
	//)

	start := currentPos.offset.End
	end := len(str)
	for start < end {
		// Iterate over nested eval expressions.
		childTok := findOpenToken(currentPos, str, start, ')')
		if childTok.token == tokenEmpty {
			// Reached EOL
			currentPos = childTok.pos
			break
		}

		// Append leftovers
		if childTok.pos.offset.Start-1 > start {
			litPos := childTok.pos
			litPos.
				strChunk := str[start:childTok.pos.offset.Start]
			children = append(children, NewLiteralExpression(litPos, strChunk))
			//se.Parts = append(se.Parts, NewLiteralExpression(NewRange(start, childTok.rng.Start-1), strChunk))
		}

		switch childTok.token {
		case tokenExprStart:
			childExpr, err := consumeExprToken(se.Pos, str, childTok)
			if err != nil {
				// Lookup possible shell expression statement end to set correct nested ranges.
				endPos := strings.IndexByte(str[childTok.rng.End:], ')')
				if endPos == -1 {
					return nil, err
				}

				err.ParentRange.EndCol = childTok.rng.End + endPos
				if isUnterminatedErr(err) {
					err.Range.EndCol = childTok.rng.End + endPos - 1
				}

				return nil, err
			}

			se.Parts = append(se.Parts, childExpr)
			start = childExpr.Range().EndCol + 1
		case tokenShellStart:
			return nil, newNestedExprError(
				ErrNestedShellExpression,
				NewRange(childTok.rng.Start, childTok.rng.End),
				se.Pos,
			)
		case tokenEnd:
			// Reached expression close
			se.Pos.EndCol = childTok.rng.End
			return se, nil
		default:
			return nil, newExprError(ErrBadToken, NewRange(childTok.rng.Start, childTok.rng.End))
		}
	}

	return nil, newNestedExprError(
		ErrUnterminatedExpression,
		se.Pos,
		parent,
	)
}
