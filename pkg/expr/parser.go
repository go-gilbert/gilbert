package expr

import (
	"errors"
	"strings"

	"github.com/go-gilbert/gilbert/pkg/parsetypes"
)

type Parser struct {
	cfg      parseConfig
	lexer    *Tokenizer
	startLoc *parsetypes.Location
}

func NewParser(src string, opts ...Option) *Parser {
	cfg := newParseConfig(opts)
	parentLoc := cfg.docInfo.location()
	parentLoc.Offset.End += len(src)

	return &Parser{
		cfg:      cfg,
		lexer:    newTokenizer(cfg, src),
		startLoc: parentLoc,
	}
}

func (p *Parser) Parse() (Expression, *parsetypes.Diagnostic) {
	expr, _, err := p.consumeUntilTok(p.startLoc, TokenTypeEmpty)
	if err != nil {
		return nil, err
	}

	if expr == nil {
		expr = NewEmptyExpression(*p.startLoc)
	}

	return expr, nil
}

func (p *Parser) consumeUntilTok(parentLoc *parsetypes.Location, endToken TokenType) (Expression, *Token, *parsetypes.Diagnostic) {
	var (
		exprs     []Expression
		strTokens []*Token
		lastTok   *Token
	)

	for tok, err := range p.lexer.IterTokens() {
		if err != nil {
			return nil, nil, p.diagFromTokenErr(err)
		}

		if tok == nil {
			break
		}

		if tok.Type == endToken {
			lastTok = tok
			break
		}

		switch tok.Type {
		case TokenTypeString, TokenTypeEOL:
			strTokens = append(strTokens, tok)
		case TokenTypeEmpty:
			panic("tokenizer returned an empty token")
		default:
			if len(strTokens) > 0 {
				exprs = append(exprs, mergeStringTokens(parentLoc, strTokens))
				strTokens = nil
			}

			expr, err := p.consumeExpression(tok)
			if err != nil {
				return nil, nil, err
			}

			exprs = append(exprs, expr)
		}
	}

	if len(strTokens) > 0 {
		exprs = append(exprs, mergeStringTokens(parentLoc, strTokens))
	}

	switch len(exprs) {
	case 0:
		return nil, lastTok, nil
	case 1:
		return exprs[0], lastTok, nil
	default:
		exp := NewCompositeExpression(*parentLoc, exprs)
		return exp, lastTok, nil
	}
}

func (p *Parser) consumeExpression(startTok *Token) (Expression, *parsetypes.Diagnostic) {
	switch startTok.Type {
	case TokenTypeEvalStart:
		return p.buildEvalExpression(startTok)
	case TokenTypeShellStart:
		return p.buildShellExpression(startTok)
	default:
		panic("unexpected non-expression token in consumeExpression")
	}
}

func (p *Parser) buildShellExpression(startTok *Token) (*ShellExpression, *parsetypes.Diagnostic) {
	loc := p.newLocationFromToken(startTok)
	body, end, err := p.consumeUntilTok(&loc, TokenTypeShellEnd)
	if err != nil {
		return nil, err
	}

	if end == nil {
		return nil, &parsetypes.Diagnostic{
			FileName: p.cfg.docInfo.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    startTok.Range,
			Offset:   parsetypes.NewOffsetRange(startTok.Offset, startTok.EndOffset()),
			Err:      errors.New("unterminated shell expression"),
		}
	}

	if body == nil {
		body = NewEmptyExpression(parsetypes.Location{
			FileName: p.cfg.docInfo.FileName,
			Offset:   parsetypes.NewOffsetRange(startTok.Offset, end.EndOffset()),
			Range:    parsetypes.NewRange(startTok.Range.Start, end.Range.End),
		})
	}

	loc.Range.End = end.Range.End
	loc.Offset.End = end.EndOffset()
	return NewShellExpression(loc, body), nil
}

func (p *Parser) buildEvalExpression(startTok *Token) (*EvalExpression, *parsetypes.Diagnostic) {
	loc := p.newLocationFromToken(startTok)
	body, end, err := p.consumeUntilTok(&loc, TokenTypeEvalEnd)
	if err != nil {
		return nil, err
	}

	if end == nil {
		return nil, &parsetypes.Diagnostic{
			FileName: p.cfg.docInfo.FileName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    startTok.Range,
			Offset:   parsetypes.NewOffsetRange(startTok.Offset, startTok.EndOffset()),
			Err:      errors.New("unterminated eval expression"),
		}
	}

	if body == nil {
		body = NewEmptyExpression(parsetypes.Location{
			FileName: p.cfg.docInfo.FileName,
			Offset:   parsetypes.NewOffsetRange(startTok.Offset, end.EndOffset()),
			Range:    parsetypes.NewRange(startTok.Range.Start, end.Range.End),
		})
	}

	loc.Range.End = end.Range.End
	loc.Offset.End = end.EndOffset()
	return NewEvalExpression(loc, body, p.cfg.evalCfg)
}

func (p *Parser) newLocationFromToken(tok *Token) parsetypes.Location {
	return parsetypes.Location{
		FileName: p.cfg.docInfo.FileName,
		Offset:   parsetypes.NewOffsetRange(tok.Offset, tok.EndOffset()),
		Range:    tok.Range,
	}
}

func (p *Parser) diagFromTokenErr(err *TokenError) *parsetypes.Diagnostic {
	return &parsetypes.Diagnostic{
		FileName: p.cfg.docInfo.FileName,
		Severity: parsetypes.DiagnosticSeverityError,
		Range:    err.Position,
		Offset:   err.Offset,
		Note:     err.Note,
		Err:      err.Err,
	}
}

func mergeStringTokens(parentLoc *parsetypes.Location, toks []*Token) Expression {
	if len(toks) == 1 {
		tok := toks[0]
		return NewStringExpression(parsetypes.Location{
			FileName: parentLoc.FileName,
			Range:    tok.Range,
			Offset: parsetypes.NewOffsetRange(
				tok.Offset, tok.EndOffset(),
			),
		}, tok.Content)
	}

	startTok := toks[0]
	endTok := toks[len(toks)-1]

	sb := &strings.Builder{}
	sb.Grow(max(len(startTok.Content), len(endTok.Content)))
	for _, tok := range toks {
		sb.WriteString(tok.Content)
	}

	return NewStringExpression(parsetypes.Location{
		FileName: parentLoc.FileName,
		Range:    parsetypes.NewRange(startTok.Range.Start, endTok.Range.End),
		Offset: parsetypes.NewOffsetRange(
			startTok.Offset, endTok.EndOffset(),
		),
	}, sb.String())
}
