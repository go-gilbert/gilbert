package expr

import (
	"errors"
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/expr-lang/expr/conf"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/stretchr/testify/require"
)

type locationInput struct {
	Offset offsetRangeString `json:"offset"`
	Range  rangeString       `json:"range"`
}

type expressionInput struct {
	Location        locationInput      `json:"location"`
	Type            string             `json:"type"`
	Value           string             `json:"value"`
	Unquote         bool               `json:"unquote"`
	ContentOffset   offsetRangeString  `json:"contentOffset"`
	ContentPosition rangeString        `json:"contentPosition"`
	Body            *expressionInput   `json:"body"`
	Parts           []*expressionInput `json:"parts"`
}

func (input *expressionInput) UnquoteValue(t *testing.T) string {
	if !input.Unquote {
		return input.Value
	}

	t.Helper()
	got, err := strconv.Unquote(`"` + input.Value + `"`)
	require.NoError(t, err, "can't unquote value")
	return got
}

type diagnostic struct {
	Location locationInput `json:"location"`
	Note     string        `json:"note"`
	Err      string        `json:"err"`
}

type parserOptsInput struct {
	Doc     *DocumentInfo `json:"doc"`
	EvalCfg *EvalConfig   `json:"evalCfg"`
}

type parserTestCase struct {
	KeepEOL bool             `json:"keepEOL"`
	Expect  *expressionInput `json:"expect"`
	Err     *diagnostic      `json:"err"`
	Opts    parserOptsInput  `json:"opts"`
}

func (tc parserTestCase) evalConfig() *EvalConfig {
	if tc.Opts.EvalCfg == nil {
		return conf.CreateNew()
	}

	return tc.Opts.EvalCfg
}

func (tc parserTestCase) options(caseName string) []Option {
	docInfo := DocumentInfo{
		FileName:      caseName,
		StartPosition: parsetypes.NewEmptyPosition(),
	}

	if tc.Opts.Doc != nil {
		docInfo = *tc.Opts.Doc
	}

	opts := []Option{
		WithDocumentInfo(docInfo),
	}

	if tc.Opts.EvalCfg != nil {
		opts = append(opts, WithEvalConfig(func(c *conf.Config) {
			*c = *tc.Opts.EvalCfg
		}))
	}

	return opts
}

type parserTestCases struct {
	Only  string                    `json:"only"`
	Cases map[string]parserTestCase `json:"cases"`
}

func (tc parserTestCase) expects(t *testing.T, caseName string) (Expression, *parsetypes.Diagnostic) {
	t.Helper()
	fname := caseName
	if tc.Opts.Doc != nil {
		fname = tc.Opts.Doc.FileName
	}

	var diag *parsetypes.Diagnostic
	if tc.Err != nil {
		diag = &parsetypes.Diagnostic{
			FileName: caseName,
			Severity: parsetypes.DiagnosticSeverityError,
			Range:    tc.Err.Location.Range.Range,
			Offset:   tc.Err.Location.Offset.Range,
			Note:     tc.Err.Note,
			Err:      errors.New(tc.Err.Err),
		}
	}

	expr := tc.mapExpression(t, fname, tc.Expect)
	return expr, diag
}

func (tc parserTestCase) mapExpression(t *testing.T, fname string, src *expressionInput) Expression {
	t.Helper()
	if src == nil {
		return nil
	}

	loc := parsetypes.Location{
		FileName: fname,
		Offset:   src.Location.Offset.Range,
		Range:    src.Location.Range.Range,
	}

	switch strings.ToLower(src.Type) {
	case "empty":
		return NewEmptyExpression(loc)
	case "string":
		return NewStringExpression(loc, src.UnquoteValue(t))
	case "shell":
		return NewShellExpression(loc, tc.mapExpression(t, fname, src.Body))
	case "composite":
		children := make([]Expression, 0, len(src.Parts))
		for _, v := range src.Parts {
			children = append(children, tc.mapExpression(t, fname, v))
		}

		return &CompositeExpression{
			header: newHeader(loc),
			Parts:  children,
		}
	case "eval":
		evalCfg := tc.evalConfig()
		expr, err := parseEvalExpr(evalCfg, src.UnquoteValue(t))
		require.NoError(t, err, "bad test input expression")

		return &EvalExpression{
			header:          newHeader(loc),
			AST:             expr,
			EvalConfig:      evalCfg,
			ContentPosition: src.ContentPosition.Range,
			ContentOffset:   src.ContentOffset.Range,
		}
	}

	t.Fatalf("invalid expression type in expects: %q", src.Type)
	return nil
}

func TestParser_Parse(t *testing.T) {
	fsys := loadTxtar(t, "parser.inputs.txtar")
	root := loadYaml[parserTestCases](t, "parser.cases.yml")

	for name, tc := range root.Cases {
		if root.Only != "" && root.Only != name {
			continue
		}

		t.Run(name, func(t *testing.T) {
			wantExprs, wantDiag := tc.expects(t, name)
			src, err := fs.ReadFile(fsys, name)
			require.NoError(t, err, "input is missing in inputs file")
			if !tc.KeepEOL {
				src = trimEOL(src)
			}

			opts := tc.options(name)
			parser := NewParser(bytesAsString(src), opts...)
			got, diag := parser.Parse()
			if wantDiag != nil {
				require.NotNil(t, diag)
				require.Equal(t, wantDiag, diag)
				return
			}

			require.Nil(t, diag)
			require.Equal(t, wantExprs, got)
		})
	}
}
