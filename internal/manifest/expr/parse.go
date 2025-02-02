package expr

import "strings"

type token int

const (
	tokenEmpty token = iota
	tokenExprStart
	tokenShellStart
	tokenEnd
)

type tokenPos struct {
	token    token
	startPos int
	endPos   int
}

func findOpenToken(str string, offset int, stopToken byte) tokenPos {
	// TODO: support escapes?
	n := len(str)

	tokenStartPos := -1
	for i := offset; i < n; i++ {
		switch v := str[i]; v {
		case '$':
			tokenStartPos = i
		case '{':
			if tokenStartPos != -1 {
				return tokenPos{
					token:    tokenExprStart,
					startPos: tokenStartPos,
					endPos:   i,
				}
			}
		case '(':
			if tokenStartPos != -1 {
				return tokenPos{
					token:    tokenShellStart,
					startPos: tokenStartPos,
					endPos:   i,
				}
			}
		default:
			tokenStartPos = -1
			if stopToken != 0 && v == stopToken {
				return tokenPos{
					token:    tokenEnd,
					startPos: i,
					endPos:   i,
				}
			}
		}
	}

	return tokenPos{
		token:    tokenEmpty,
		startPos: offset,
		endPos:   len(str) - 1,
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
func Parse(str string) (Expression, error) {
	if str == "" {
		return EmptyExpression{}, nil
	}

	rootRng := NewRange(0, len(str)-1)
	root := NewCompositeExpression(
		NewRange(0, len(str)-1),
		make([]Expression, 0, 1),
	)

	start := 0
	end := len(str)
	for start < end {
		tok := findOpenToken(str, start, 0)
		if tok.token == tokenEmpty {
			// Break if nothing to parse
			strChunk := str[start : tok.endPos+1]
			root.Parts = append(
				root.Parts,
				NewLiteralExpression(NewRange(start, tok.endPos), strChunk),
			)
			break
		}

		// Append adjacent string literal if any.
		if tok.startPos-1 > start {
			strChunk := str[start:tok.startPos]
			root.Parts = append(
				root.Parts,
				NewLiteralExpression(NewRange(start, tok.startPos-1), strChunk),
			)
		}

		expr, err := consumeToken(rootRng, str, tok)
		if err != nil {
			return nil, err
		}

		root.Parts = append(root.Parts, expr)
		start = expr.Range().EndCol + 1
	}

	// Unwrap if there is only one statement
	if len(root.Parts) == 1 {
		return root.Parts[0], nil
	}

	return root, nil
}

func consumeToken(parent Range, str string, pos tokenPos) (Expression, *ExpressionError) {
	switch pos.token {
	case tokenExprStart:
		return consumeExprToken(parent, str, pos)
	case tokenShellStart:
		return consumeShellToken(parent, str, pos)
	default:
		return nil, newNestedExprError(ErrBadToken, NewRange(pos.startPos, pos.endPos), parent)
	}
}

func consumeExprToken(parent Range, str string, pos tokenPos) (Expression, *ExpressionError) {
	endPos := -1
	for i := pos.endPos + 1; i < len(str); i++ {
		if str[i] == '}' {
			endPos = i
			break
		}
	}

	if endPos == -1 {
		return nil, newNestedExprError(
			ErrUnterminatedExpression,
			NewRange(pos.startPos, len(str)-1),
			parent,
		)
	}

	tokenRng := NewRange(pos.startPos, endPos)
	content := str[pos.endPos+1 : endPos]
	if content == "" {
		return nil, newNestedExprError(
			ErrEmptyExpression,
			tokenRng,
			parent,
		)
	}

	exp, err := NewEvalExpression(tokenRng, content, nil)
	if err != nil {
		return nil, newNestedExprError(err, tokenRng, parent)
	}

	return exp, nil
}

func consumeShellToken(parent Range, str string, pos tokenPos) (Expression, *ExpressionError) {
	se := NewShellExpression(
		NewRange(pos.startPos, parent.EndCol),
		make([]Expression, 0, 1),
	)

	start := pos.endPos + 1
	end := len(str)
	for start < end {
		// Iterate over nested eval expressions.
		tok := findOpenToken(str, start, ')')
		if tok.token == tokenEmpty {
			// Reached EOL
			break
		}

		// Append leftovers
		if tok.startPos-1 > start {
			strChunk := str[start:tok.startPos]
			se.Parts = append(se.Parts, NewLiteralExpression(NewRange(start, tok.startPos-1), strChunk))
		}

		switch tok.token {
		case tokenExprStart:
			childExpr, err := consumeExprToken(se.Pos, str, tok)
			if err != nil {
				// Lookup possible shell expression statement end to set correct nested ranges.
				endPos := strings.IndexByte(str[tok.endPos:], ')')
				if endPos == -1 {
					return nil, err
				}

				err.ParentRange.EndCol = tok.endPos + endPos
				if isUnterminatedErr(err) {
					err.Range.EndCol = tok.endPos + endPos - 1
				}

				return nil, err
			}

			se.Parts = append(se.Parts, childExpr)
			start = childExpr.Range().EndCol + 1
		case tokenShellStart:
			return nil, newNestedExprError(
				ErrNestedShellExpression,
				NewRange(tok.startPos, tok.endPos),
				se.Pos,
			)
		case tokenEnd:
			// Reached expression close
			se.Pos.EndCol = tok.endPos
			return se, nil
		default:
			return nil, newExprError(ErrBadToken, NewRange(tok.startPos, tok.endPos))
		}
	}

	return nil, newNestedExprError(
		ErrUnterminatedExpression,
		se.Pos,
		parent,
	)
}
