// package functions contains built-in functions
// that can be used in hcl manifest file
package context

import (
	"fmt"
	"github.com/go-gilbert/gilbert/support/shell"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"os"
	"strings"
)

func GetDefaultFunctions() map[string]function.Function {
	return map[string]function.Function{
		"split": SplitFunc,
		"shell": ShellFunc,
	}
}

// SplitFunc splits string by delimiter and returns array of strings
var SplitFunc = function.New(&function.Spec{
	Type: function.StaticReturnType(cty.List(cty.String)),
	Params: []function.Parameter{
		{Name: "value", Type: cty.String},
		{Name: "delimiter", Type: cty.String},
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		// TODO: add type check
		if len(args) == 0 {
			return cty.NilVal, fmt.Errorf("expected 2 arguments but got 0")
		}

		value := args[0].AsString()
		delimiter := args[1].AsString()
		result := strings.Split(value, delimiter)

		values := make([]cty.Value, 0, len(result))
		for _, v := range result {
			values = append(values, cty.StringVal(v))
		}

		return cty.ListVal(values), nil
	},
})

// ShellFunc runs shell command and returns value
var ShellFunc = function.New(&function.Spec{
	Type: function.StaticReturnType(cty.String),
	Params: []function.Parameter{
		{Name: "command", Type: cty.String},
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		// TODO: add type check
		if len(args) == 0 {
			return cty.NilVal, fmt.Errorf("expected 1 argument but got 0")
		}

		command := args[0].AsString()
		cmd := shell.PrepareCommand(command)
		cmd.Env = os.Environ()
		data, err := cmd.CombinedOutput()
		if err != nil {
			return cty.NilVal, fmt.Errorf("%s (%s)", shell.FormatExitError(err), data)
		}

		return cty.StringVal(string(data)), nil
	},
})
