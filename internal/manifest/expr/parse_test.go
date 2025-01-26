package expr

import (
	"testing"

	"github.com/expr-lang/expr/conf"
	"github.com/stretchr/testify/require"
)

func syntaxErrFromExpr(exp string) error {
	_, err := parseEvalExpr(nil, exp)
	return err
}

func mustNewEvalExpr(t *testing.T, pos Range, str string, cfg *conf.Config) *EvalExpression {
	t.Helper()
	exp, err := NewEvalExpression(pos, str, cfg)
	require.NoError(t, err)
	return exp
}

func TestParse(t *testing.T) {
	// TODO: fuzzing
	cases := []struct {
		label   string
		input   string
		wantFn  func(input string, t *testing.T) Expression
		wantErr error
	}{
		{
			label: "string literal only",
			input: "hello world",
			wantFn: func(input string, t *testing.T) Expression {
				return NewLiteralExpression(NewRange(0, 10), "hello world")
			},
		},
		{
			label: "expression only",
			input: "${foobar}",
			wantFn: func(input string, t *testing.T) Expression {
				return mustNewEvalExpr(t, NewRange(0, 8), "foobar", nil)
			},
		},
		{
			label: "mix of literal and expression",
			input: "foo ${bar} fizz${baz}?",
			wantFn: func(input string, t *testing.T) Expression {
				return NewCompositeExpression(
					NewRange(0, len(input)-1),
					[]Expression{
						NewLiteralExpression(NewRange(0, 3), "foo "),
						mustNewEvalExpr(t, NewRange(4, 9), "bar", nil),
						NewLiteralExpression(NewRange(10, 14), " fizz"),
						mustNewEvalExpr(t, NewRange(15, 20), "baz", nil),
						NewLiteralExpression(NewRange(21, 21), "?"),
					},
				)
			},
		},
		{
			label: "mix of shell and literal",
			input: "$(fizz)buzz",
			wantFn: func(input string, t *testing.T) Expression {
				return NewCompositeExpression(
					NewRange(0, len(input)-1),
					[]Expression{
						NewShellExpression(NewRange(0, 6), []Expression{
							NewLiteralExpression(NewRange(2, 5), "fizz"),
						}),
						NewLiteralExpression(NewRange(7, 10), "buzz"),
					},
				)
			},
		},
		{
			label: "mix of literal, shell and expression",
			input: "foo ${bar} fizz$(whoami)?",
			wantFn: func(input string, t *testing.T) Expression {
				return NewCompositeExpression(
					NewRange(0, len(input)-1),
					[]Expression{
						NewLiteralExpression(NewRange(0, 3), "foo "),
						mustNewEvalExpr(t, NewRange(4, 9), "bar", nil),
						NewLiteralExpression(NewRange(10, 14), " fizz"),
						NewShellExpression(NewRange(15, 23), []Expression{
							NewLiteralExpression(NewRange(17, 22), "whoami"),
						}),
						NewLiteralExpression(NewRange(24, 24), "?"),
					},
				)
			},
		},
		{
			label: "shell only",
			input: "$(ls -la)",
			wantFn: func(input string, t *testing.T) Expression {
				return NewShellExpression(
					NewRange(0, 8),
					[]Expression{
						NewLiteralExpression(NewRange(2, 7), "ls -la"),
					},
				)
			},
		},
		{
			label: "shell with expression",
			input: "$(chown ${env.USER} 777)",
			wantFn: func(input string, t *testing.T) Expression {
				return NewShellExpression(
					NewRange(0, 23),
					[]Expression{
						NewLiteralExpression(NewRange(2, 7), "chown "),
						mustNewEvalExpr(t, NewRange(8, 18), "env.USER", nil),
						NewLiteralExpression(NewRange(19, 22), " 777"),
					},
				)
			},
		},
		{
			label: "shell with parentheses expression",
			input: "$(foo ${4 + (2*2)} bar)",
			wantFn: func(input string, t *testing.T) Expression {
				return NewShellExpression(
					NewRange(0, len(input)-1),
					[]Expression{
						NewLiteralExpression(NewRange(2, 5), "foo "),
						mustNewEvalExpr(t, NewRange(6, 17), "4 + (2*2)", nil),
						NewLiteralExpression(NewRange(18, 21), " bar"),
					},
				)
			},
		},

		// Errors
		{
			label: "unterminated expression",
			input: "foo ${bar",
			wantErr: newNestedExprError(
				ErrUnterminatedExpression,
				NewRange(4, 8),
				NewRange(0, 8),
			),
		},
		{
			label: "unterminated shell expression",
			input: "foo $(bar",
			wantErr: newNestedExprError(
				ErrUnterminatedExpression,
				NewRange(4, 8),
				NewRange(0, 8),
			),
		},
		{
			label: "empty expression",
			input: "foo ${}",
			wantErr: newNestedExprError(
				ErrEmptyExpression,
				NewRange(4, 6),
				NewRange(0, 6),
			),
		},
		{
			label: "forbidden nested shell expr",
			input: "foo $(foo $(bar))",
			wantErr: newNestedExprError(
				ErrNestedShellExpression,
				NewRange(10, 11),
				NewRange(4, 16),
			),
		},
		{
			label: "nested unterminated expr",
			input: "foo $(foo ${bar",
			wantErr: newNestedExprError(
				ErrUnterminatedExpression,
				NewRange(10, 14),
				NewRange(4, 14),
			),
		},
		{
			label: "nested unterminated expr with bounds",
			input: "foo $(foo ${bar)",
			wantErr: newNestedExprError(
				ErrUnterminatedExpression,
				NewRange(10, 14),
				NewRange(4, 15),
			),
		},
		{
			label: "nested unterminated expr in middle",
			input: "foo $(foo ${bar) baz",
			wantErr: newNestedExprError(
				ErrUnterminatedExpression,
				NewRange(10, 14),
				NewRange(4, 15),
			),
		},
		{
			label: "syntax error",
			input: "foo $(foo ${bar/}) baz",
			wantErr: newNestedExprError(
				syntaxErrFromExpr("bar/"),
				NewRange(10, 16),
				NewRange(4, 17),
			),
		},
		{
			label: "empty string",
			input: "",
			wantFn: func(input string, t *testing.T) Expression {
				return EmptyExpression{}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			got, err := Parse(c.input)
			if c.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, c.wantErr, err)
				return
			}

			want := c.wantFn(c.input, t)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}
