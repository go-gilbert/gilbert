package expr

import (
	"testing"

	"github.com/go-gilbert/gilbert/internal/manifest/expr/exprmock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -package=exprmock -destination=exprmock/context.go github.com/go-gilbert/gilbert/internal/manifest/expr CommandProcessor,ValueResolver
func TestParser_ReadString(t *testing.T) {
	cases := map[string]struct {
		input      string
		ctx        EvalContext
		expect     string
		expectErr  string
		skip       bool
		getContext func(t *testing.T, ctrl *gomock.Controller) EvalContext
	}{
		"expand variable": {
			input:  "${foo}",
			expect: "bar",
			getContext: func(t *testing.T, ctrl *gomock.Controller) EvalContext {
				valRes := exprmock.NewMockValueResolver(ctrl)
				valRes.EXPECT().ValueByName("foo").Return("bar", true)
				return EvalContext{
					Env: valRes,
				}
			},
		},
		"replace variable in string": {
			input:  "result is ${foo} = ${bar}!",
			expect: "result is 2+2 = 4!",
			getContext: func(t *testing.T, ctrl *gomock.Controller) EvalContext {
				valRes := exprmock.NewMockValueResolver(ctrl)
				valRes.EXPECT().ValueByName("foo").Return("2+2", true)
				valRes.EXPECT().ValueByName("bar").Return("4", true)
				return EvalContext{
					Env: valRes,
				}
			},
		},
		"evaluate shell expressions": {
			input:  "$(ls -la)",
			expect: "foo.go",
			getContext: func(t *testing.T, ctrl *gomock.Controller) EvalContext {
				cmdProc := exprmock.NewMockCommandProcessor(ctrl)
				cmdProc.EXPECT().EvalCommand("ls -la").Return([]byte("foo.go"), nil)
				return EvalContext{
					CommandProcessor: cmdProc,
				}
			},
		},
		"expand both vars and commands": {
			input:  "result of command ${ cmdname } is $( custom command )",
			expect: "result of command uname -sm is Linux aarch64",
			getContext: func(t *testing.T, ctrl *gomock.Controller) EvalContext {
				cmdProc := exprmock.NewMockCommandProcessor(ctrl)
				cmdProc.EXPECT().EvalCommand("custom command").Return([]byte("Linux aarch64"), nil)

				varRes := exprmock.NewMockValueResolver(ctrl)
				varRes.EXPECT().ValueByName("cmdname").Return("uname -sm", true)

				return EvalContext{
					CommandProcessor: cmdProc,
					Env:              varRes,
				}
			},
		},
		"var is undefined": {
			input:     "${foo.bar}",
			expectErr: `"foo.bar" is not defined`,
			getContext: func(t *testing.T, ctrl *gomock.Controller) EvalContext {
				valRes := exprmock.NewMockValueResolver(ctrl)
				valRes.EXPECT().ValueByName("foo.bar").Return("", false)

				return EvalContext{
					Env: valRes,
				}
			},
		},
		"handle plain strings": {
			input:  "foo bar baz",
			expect: "foo bar baz",
			getContext: func(t *testing.T, ctrl *gomock.Controller) EvalContext {
				return EvalContext{}
			},
		},
		"parse vars inside shell expressions": {
			skip:   true,
			input:  "result of command ${ cmdname } is $( ${ cmdname } )",
			expect: "result of command uname -sm is Linux aarch64",
			getContext: func(t *testing.T, ctrl *gomock.Controller) EvalContext {
				cmdProc := exprmock.NewMockCommandProcessor(ctrl)
				cmdProc.EXPECT().EvalCommand("uname -sm").Return([]byte("Linux aarch64"), nil)

				varRes := exprmock.NewMockValueResolver(ctrl)
				varRes.EXPECT().ValueByName("cmdname").Return("uname -sm", true)

				return EvalContext{
					CommandProcessor: cmdProc,
					Env:              varRes,
				}
			},
		},
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			if c.skip {
				t.Skip("TODO")
				return
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := c.getContext(t, ctrl)
			p := NewSpecV2Parser()
			got, err := p.ReadString(ctx, c.input)
			if c.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.expectErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, c.expect, got)
		})
	}
}
