package expr

import (
	"errors"
	"testing"

	"github.com/expr-lang/expr/file"
	"github.com/go-gilbert/gilbert/internal/manifest/expr/exprmock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEmptyExpression(t *testing.T) {
	var exp EmptyExpression
	require.False(t, exp.Evaluable())
	require.Equal(t, exp.Range(), Range{})

	b, err := exp.Eval(EvalContext{})
	require.NoError(t, err)
	require.Empty(t, b)

	v, err := exp.String(EvalContext{})
	require.NoError(t, err)
	require.Empty(t, v)
}

func TestLiteralExpression(t *testing.T) {
	val := "foobar"
	rng := NewRange(0, len(val)-1)
	exp := NewLiteralExpression(rng, val)

	require.False(t, exp.Evaluable())
	require.Equal(t, exp.Range(), rng)

	b, err := exp.Eval(EvalContext{})
	require.NoError(t, err)
	require.Equal(t, val, b)

	v, err := exp.String(EvalContext{})
	require.NoError(t, err)
	require.Equal(t, []byte(val), v)
}

func TestEvalExpression_Evaluable(t *testing.T) {
	var exp EvalExpression
	require.True(t, exp.Evaluable())
}

func TestEvalExpression_Eval(t *testing.T) {
	cases := []struct {
		name      string
		inputFn   func(t *testing.T) *EvalExpression
		contextFn func(ctrl *gomock.Controller) EvalContext
		want      any
		wantErr   error
	}{
		{
			name: "correct value",
			want: 4,
			inputFn: func(t *testing.T) *EvalExpression {
				v, err := NewEvalExpression(Range{}, "2 + x", nil)
				require.NoError(t, err)
				return v
			},
			contextFn: func(ctrl *gomock.Controller) EvalContext {
				vr := exprmock.NewMockValueResolver(ctrl)
				vr.EXPECT().Values().Return(map[string]any{"x": 2})

				return EvalContext{
					Env: vr,
				}
			},
		},
		{
			name: "runtime error",
			wantErr: newExprError(
				&file.Error{
					Location: file.Location{
						From: 2,
						To:   3,
					},
					Line:    1,
					Column:  1,
					Message: "invalid operation: int + <nil>",
					Snippet: "\n | 2 +x\n | ..^",
				},
				NewRange(0, 10),
			),
			inputFn: func(t *testing.T) *EvalExpression {
				v, err := NewEvalExpression(NewRange(0, 10), "2 + x", nil)
				require.NoError(t, err)
				return v
			},
			contextFn: func(ctrl *gomock.Controller) EvalContext {
				vr := exprmock.NewMockValueResolver(ctrl)
				vr.EXPECT().Values().Return(map[string]any{})

				return EvalContext{
					Env: vr,
				}
			},
		},
		{
			name: "syntax error",
			wantErr: newExprError(
				errors.New("runtime error: invalid memory address or nil pointer dereference"),
				NewRange(0, 10),
			),
			inputFn: func(t *testing.T) *EvalExpression {
				v, err := NewEvalExpression(NewRange(0, 10), "2 + x", nil)
				require.NoError(t, err)
				v.AST = nil
				v.EvalConfig = nil
				return v
			},
			contextFn: func(ctrl *gomock.Controller) EvalContext {
				vr := exprmock.NewMockValueResolver(ctrl)
				return EvalContext{
					Env: vr,
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			input := c.inputFn(t)
			ctx := c.contextFn(ctrl)

			got, err := input.Eval(ctx)
			if c.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, c.want, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, c.want, got)
		})
	}
}

func TestEvalExpression_String(t *testing.T) {
	cases := []struct {
		name      string
		inputFn   func(t *testing.T) *EvalExpression
		contextFn func(ctrl *gomock.Controller) EvalContext
		want      string
		wantErr   error
	}{
		{
			name: "return string",
			want: "foobar",
			inputFn: func(t *testing.T) *EvalExpression {
				v, err := NewEvalExpression(NewRange(0, 10), "x", nil)
				require.NoError(t, err)
				return v
			},
			contextFn: func(ctrl *gomock.Controller) EvalContext {
				vr := exprmock.NewMockValueResolver(ctrl)
				vr.EXPECT().Values().Return(map[string]any{
					"x": "foobar",
				})

				return EvalContext{
					Env: vr,
				}
			},
		},
		{
			name: "cast bool to string",
			want: "true",
			inputFn: func(t *testing.T) *EvalExpression {
				v, err := NewEvalExpression(NewRange(0, 10), "x", nil)
				require.NoError(t, err)
				return v
			},
			contextFn: func(ctrl *gomock.Controller) EvalContext {
				vr := exprmock.NewMockValueResolver(ctrl)
				vr.EXPECT().Values().Return(map[string]any{
					"x": true,
				})

				return EvalContext{
					Env: vr,
				}
			},
		},
		{
			name: "cast number to string",
			want: "1234",
			inputFn: func(t *testing.T) *EvalExpression {
				v, err := NewEvalExpression(NewRange(0, 10), "x", nil)
				require.NoError(t, err)
				return v
			},
			contextFn: func(ctrl *gomock.Controller) EvalContext {
				vr := exprmock.NewMockValueResolver(ctrl)
				vr.EXPECT().Values().Return(map[string]any{
					"x": 1234,
				})

				return EvalContext{
					Env: vr,
				}
			},
		},
		{
			name: "throw error on non-primitive values",
			wantErr: newExprError(
				errors.New("value struct {}{} cannot be converted to a string"),
				NewRange(0, 10),
			),
			inputFn: func(t *testing.T) *EvalExpression {
				v, err := NewEvalExpression(NewRange(0, 10), "x", nil)
				require.NoError(t, err)
				return v
			},
			contextFn: func(ctrl *gomock.Controller) EvalContext {
				vr := exprmock.NewMockValueResolver(ctrl)
				vr.EXPECT().Values().Return(map[string]any{
					"x": struct{}{},
				})

				return EvalContext{
					Env: vr,
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			input := c.inputFn(t)
			ctx := c.contextFn(ctrl)

			got, err := input.String(ctx)
			if c.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, c.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, []byte(c.want), got)
		})
	}
}

func TestShellExpression_Evaluable(t *testing.T) {
	var e ShellExpression
	require.True(t, e.Evaluable())
}

func TestShellExpression_Eval(t *testing.T) {
	cases := []struct {
		label   string
		exprFn  func(t *testing.T) *ShellExpression
		ctxFn   func(ctrl *gomock.Controller) EvalContext
		want    string
		wantErr error
	}{
		{
			label: "call command and return result",
			want:  "arm64",
			exprFn: func(t *testing.T) *ShellExpression {
				return NewShellExpression(
					Range{},
					[]Expression{
						NewLiteralExpression(Range{}, " uname -m "),
					},
				)
			},
			ctxFn: func(ctrl *gomock.Controller) EvalContext {
				cmdProc := exprmock.NewMockCommandProcessor(ctrl)
				cmdProc.EXPECT().EvalCommand("uname -m").Return([]byte("arm64"), nil)
				return EvalContext{
					CommandProcessor: cmdProc,
				}
			},
		},
		{
			label:   "return command error",
			wantErr: newExprError(errors.New("foobar"), NewRange(5, 10)),
			exprFn: func(t *testing.T) *ShellExpression {
				return NewShellExpression(
					NewRange(5, 10),
					[]Expression{
						NewLiteralExpression(Range{}, "foobar"),
					},
				)
			},
			ctxFn: func(ctrl *gomock.Controller) EvalContext {
				cmdProc := exprmock.NewMockCommandProcessor(ctrl)
				cmdProc.EXPECT().EvalCommand("foobar").Return(nil, errors.New("foobar"))
				return EvalContext{
					CommandProcessor: cmdProc,
				}
			},
		},
		{
			label: "return expand error",
			wantErr: newNestedExprError(
				&file.Error{
					Location: file.Location{
						From: 0,
						To:   1,
					},
					Line:    1,
					Column:  0,
					Message: "cannot fetch x from <nil>",
					Snippet: "\n | x\n | ^",
				},
				NewRange(5, 9), NewRange(0, 10),
			),
			exprFn: func(t *testing.T) *ShellExpression {
				return NewShellExpression(
					NewRange(0, 10),
					[]Expression{
						NewLiteralExpression(NewRange(0, 4), "foobar"),
						must(NewEvalExpression(
							NewRange(5, 9), "x", nil,
						)),
					},
				)
			},
			ctxFn: func(ctrl *gomock.Controller) EvalContext {
				env := exprmock.NewMockValueResolver(ctrl)
				env.EXPECT().Values().Return(nil)

				return EvalContext{
					Env: env,
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			expr := c.exprFn(t)
			ctx := c.ctxFn(ctrl)
			got, err := expr.Eval(ctx)
			if c.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, c.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, c.want, got)
		})
	}
}

func TestCompositeExpression_Eval(t *testing.T) {
	cases := []struct {
		label   string
		exprFn  func(t *testing.T) *CompositeExpression
		ctxFn   func(ctrl *gomock.Controller) EvalContext
		want    string
		wantErr error
	}{
		{
			label: "return result",
			want:  "foobar",
			exprFn: func(t *testing.T) *CompositeExpression {
				return NewCompositeExpression(
					Range{},
					[]Expression{
						NewLiteralExpression(Range{}, "foobar"),
					},
				)
			},
			ctxFn: func(ctrl *gomock.Controller) EvalContext {
				return EvalContext{}
			},
		},
		{
			label: "return command error",
			wantErr: newNestedExprError(
				errors.New("foobar"), NewRange(1, 9), NewRange(0, 10),
			),
			exprFn: func(t *testing.T) *CompositeExpression {
				return NewCompositeExpression(
					NewRange(0, 10),
					[]Expression{
						NewShellExpression(NewRange(1, 9),
							[]Expression{
								NewLiteralExpression(NewRange(2, 8), "cmd"),
							},
						),
					},
				)
			},
			ctxFn: func(ctrl *gomock.Controller) EvalContext {
				cmdProc := exprmock.NewMockCommandProcessor(ctrl)
				cmdProc.EXPECT().EvalCommand("cmd").Return(nil, errors.New("foobar"))
				return EvalContext{
					CommandProcessor: cmdProc,
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			expr := c.exprFn(t)
			ctx := c.ctxFn(ctrl)
			got, err := expr.Eval(ctx)
			if c.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, c.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, []byte(c.want), got)
		})
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
