package expr

import (
	"testing"

	"github.com/expr-lang/expr/conf"
	"github.com/stretchr/testify/require"
)

func mustNewEvalExpr(t *testing.T, pos Range, str string, cfg *conf.Config) *EvalExpression {
	t.Helper()
	exp, err := NewEvalExpression(pos, str, cfg)
	require.NoError(t, err)
	return exp
}

func TestParse(t *testing.T) {
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
