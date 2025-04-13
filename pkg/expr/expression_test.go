package expr

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-gilbert/gilbert/pkg/expr/exprmock"
	"github.com/go-gilbert/gilbert/pkg/parsetypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type Record = map[string]any

//go:generate mockgen -destination ./exprmock/eval.go -package exprmock . CommandProcessor,ValueResolver
func TestExpression_Eval(t *testing.T) {
	cases := []struct {
		name       string
		src        string
		evalParams func(ctrl *gomock.Controller) EvalParams
		want       any
		wantDiag   *parsetypes.Diagnostic
	}{
		{
			name: "should run shell with expanded vars",
			src:  "$(id ${{env.USER}})",
			want: "result!",
			evalParams: buildEvalParams(func(cmdProc *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				valRes.EXPECT().Values().Return(Record{
					"env": Record{
						"USER": "root",
					},
				}, nil)
				cmdProc.EXPECT().
					EvalCommand(gomock.Any(), "id root").
					Return([]byte("result!"), nil)
			}),
		},
		{
			name: "should expand eval in string",
			src:  "answer is ${{answer}}",
			want: "answer is 42",
			evalParams: buildEvalParams(func(_ *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				valRes.EXPECT().Values().Return(Record{
					"answer": 42,
				}, nil)
			}),
		},
		{
			name: "empty string should be evaluable",
			src:  "",
			want: nil,
		},
		{
			name: "should wrap eval error into diagnostics",
			src:  "${{foo.bar}}",
			evalParams: buildEvalParams(func(_ *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				valRes.EXPECT().Values().Return(Record{}, nil)
			}),
			wantDiag: &parsetypes.Diagnostic{
				Severity: parsetypes.DiagnosticSeverityError,
				Range: parsetypes.NewRange(
					parsetypes.NewPosition(1, 7),
					parsetypes.NewPosition(1, 10),
				),
				Offset: parsetypes.NewOffsetRange(6, 9),
				Note:   "cannot fetch bar from <nil>",
				Err:    errors.New("expression error: cannot fetch bar from <nil>"),
			},
		},
		{
			name: "should wrap shell error into diagnostic",
			src:  "$(test)",
			evalParams: buildEvalParams(func(cmdProc *exprmock.MockCommandProcessor, _ *exprmock.MockValueResolver) {
				cmdProc.EXPECT().EvalCommand(gomock.Any(), "test").
					Return(nil, errors.New("err msg"))
			}),
			wantDiag: &parsetypes.Diagnostic{
				Severity: parsetypes.DiagnosticSeverityError,
				Range: parsetypes.NewRange(
					parsetypes.NewPosition(1, 3),
					parsetypes.NewPosition(1, 6),
				),
				Offset: parsetypes.NewOffsetRange(2, 5),
				Err:    fmt.Errorf("shell expression error: %w", errors.New("err msg")),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exp, err := NewParser(tc.src).Parse()
			require.Nil(t, err, "test input has errors")

			evParams := NoopEvalParams
			if tc.evalParams != nil {
				ctrl := gomock.NewController(t)
				evParams = tc.evalParams(ctrl)
			}

			got, err := exp.Eval(t.Context(), evParams)
			if tc.wantDiag != nil {
				require.NotNil(t, err, "expected error but got nil")
				require.Equal(t, tc.wantDiag, err)
				return
			}

			require.Nil(t, err, "unexpected error")
			require.Equal(t, tc.want, got)
		})
	}
}

func TestExpression_EvalText(t *testing.T) {
	cases := []struct {
		name       string
		src        string
		evalParams func(ctrl *gomock.Controller) EvalParams
		want       string
		wantDiag   *parsetypes.Diagnostic
	}{
		{
			name: "should run shell with expanded vars",
			src:  "$(id ${{env.USER}})",
			want: "result!",
			evalParams: buildEvalParams(func(cmdProc *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				valRes.EXPECT().Values().Return(Record{
					"env": Record{
						"USER": "root",
					},
				}, nil)
				cmdProc.EXPECT().
					EvalCommand(gomock.Any(), "id root").
					Return([]byte("result!"), nil)
			}),
		},
		{
			name: "should cast bool to string",
			src:  "${{v}}",
			want: "true",
			evalParams: buildEvalParams(func(_ *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				valRes.EXPECT().Values().Return(Record{
					"v": true,
				}, nil)
			}),
		},
		{
			name: "should cast number to string",
			src:  "${{answer}}",
			want: "42",
			evalParams: buildEvalParams(func(_ *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				valRes.EXPECT().Values().Return(Record{
					"answer": 42,
				}, nil)
			}),
		},
		{
			name: "should cast printable to string",
			src:  "val: ${{v}}",
			want: "val: 1991-08-24 12:30:00 +0300 EEST",
			evalParams: buildEvalParams(func(_ *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				tz, err := time.LoadLocation("Europe/Kiev")
				if err != nil {
					panic(err)
				}

				valRes.EXPECT().Values().Return(Record{
					"v": time.Date(1991, 8, 24, 12, 30, 0, 0, tz),
				}, nil)
			}),
		},
		{
			name: "should expand eval in string",
			src:  "answer is ${{answer}}",
			want: "answer is 42",
			evalParams: buildEvalParams(func(_ *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				valRes.EXPECT().Values().Return(Record{
					"answer": 42,
				}, nil)
			}),
		},
		{
			name: "empty string should be evaluable",
			src:  "",
			want: "",
		},
		{
			name: "should report error for uncastable values",
			src:  "${{v}}",
			evalParams: buildEvalParams(func(_ *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				valRes.EXPECT().Values().Return(Record{
					"v": Record{},
				}, nil)
			}),
			wantDiag: &parsetypes.Diagnostic{
				Severity: parsetypes.DiagnosticSeverityError,
				Range: parsetypes.NewRange(
					parsetypes.NewPosition(1, 4),
					parsetypes.NewPosition(1, 4),
				),
				Offset: parsetypes.NewOffsetRange(3, 3),
				Err:    errors.New("value of type map[string]interface {} cannot be converted to a string"),
			},
		},
		{
			name: "should wrap eval error into diagnostics",
			src:  "${{foo.bar}}",
			evalParams: buildEvalParams(func(_ *exprmock.MockCommandProcessor, valRes *exprmock.MockValueResolver) {
				valRes.EXPECT().Values().Return(Record{}, nil)
			}),
			wantDiag: &parsetypes.Diagnostic{
				Severity: parsetypes.DiagnosticSeverityError,
				Range: parsetypes.NewRange(
					parsetypes.NewPosition(1, 7),
					parsetypes.NewPosition(1, 10),
				),
				Offset: parsetypes.NewOffsetRange(6, 9),
				Note:   "cannot fetch bar from <nil>",
				Err:    errors.New("expression error: cannot fetch bar from <nil>"),
			},
		},
		{
			name: "should wrap shell error into diagnostic",
			src:  "$(test)",
			evalParams: buildEvalParams(func(cmdProc *exprmock.MockCommandProcessor, _ *exprmock.MockValueResolver) {
				cmdProc.EXPECT().EvalCommand(gomock.Any(), "test").
					Return(nil, errors.New("err msg"))
			}),
			wantDiag: &parsetypes.Diagnostic{
				Severity: parsetypes.DiagnosticSeverityError,
				Range: parsetypes.NewRange(
					parsetypes.NewPosition(1, 3),
					parsetypes.NewPosition(1, 6),
				),
				Offset: parsetypes.NewOffsetRange(2, 5),
				Err:    fmt.Errorf("shell expression error: %w", errors.New("err msg")),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exp, err := NewParser(tc.src).Parse()
			require.Nil(t, err, "test input has errors")

			evParams := NoopEvalParams
			if tc.evalParams != nil {
				ctrl := gomock.NewController(t)
				evParams = tc.evalParams(ctrl)
			}

			got, err := exp.EvalText(t.Context(), evParams)
			if tc.wantDiag != nil {
				require.NotNil(t, err, "expected error but got nil")
				require.Equal(t, tc.wantDiag, err)
				return
			}

			require.Nil(t, err, "unexpected error")
			require.Equal(t, tc.want, string(got))
		})
	}
}

func buildEvalParams(fn func(*exprmock.MockCommandProcessor, *exprmock.MockValueResolver)) func(*gomock.Controller) EvalParams {
	return func(ctrl *gomock.Controller) EvalParams {
		valRes := exprmock.NewMockValueResolver(ctrl)
		cmdProc := exprmock.NewMockCommandProcessor(ctrl)
		fn(cmdProc, valRes)

		return EvalParams{
			CommandProcessor: cmdProc,
			Env:              valRes,
		}
	}
}

func TestExpression_Evaluable(t *testing.T) {
	cases := []struct {
		name string
		expr Expression
		want bool
	}{
		{
			name: "empty",
			want: false,
			expr: &EmptyExpression{},
		},
		{
			name: "composite-static",
			want: false,
			expr: &CompositeExpression{
				Parts: []Expression{
					&StringExpression{},
				},
			},
		},
		{
			name: "composite-dynamic",
			want: true,
			expr: &CompositeExpression{
				Parts: []Expression{
					&EvalExpression{},
					&StringExpression{},
				},
			},
		},
		{
			name: "eval",
			want: true,
			expr: &EvalExpression{},
		},
		{
			name: "shell-empty",
			want: false,
			expr: &ShellExpression{},
		},
		{
			name: "shell-filled",
			want: true,
			expr: &ShellExpression{
				Body: &StringExpression{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.expr.Evaluable()
			require.Equal(t, tc.want, got)
		})
	}
}
