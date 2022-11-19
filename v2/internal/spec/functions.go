package spec

import (
	"bytes"
	"fmt"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// TODO: generate cty functions from Go source file
var builtinFunctions = map[string]function.Function{
	"shell": function.New(&function.Spec{
		Description: "Call shell command and return a result",
		Params: []function.Parameter{
			{
				Name: "command",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			cmdStr := strings.TrimSpace(args[0].AsString())
			chunks := strings.Split(cmdStr, " ")

			cmd := exec.Command(chunks[0])
			cmd.Args = chunks

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			cmd.Stdout = stdout
			cmd.Stderr = stderr

			if err := cmd.Start(); err != nil {
				return cty.NilVal, fmt.Errorf("failed to start command: %w", err)
			}

			if err := cmd.Wait(); err != nil {
				if stderr.Len() == 0 {
					return cty.NilVal, fmt.Errorf("command returned an error: %w", err)
				}

				return cty.NilVal, fmt.Errorf("command returned at error: %s (%w)", stderr.String(), err)
			}

			return cty.StringVal(stdout.String()), nil
		},
	}),

	"regex_match": function.New(&function.Spec{
		Description: "Check if string matches regular expression",
		Params: []function.Parameter{
			{Name: "pattern", Type: cty.String},
			{Name: "string", Type: cty.String},
		},
		VarParam: nil,
		Type:     function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			matched, err := regexp.MatchString(args[1].AsString(), args[0].AsString())
			if err != nil {
				return cty.NilVal, err
			}

			return cty.BoolVal(matched), err
		},
	}),

	"basename": function.New(&function.Spec{
		Description: "Returns base file name from file path",
		Params: []function.Parameter{
			{Name: "filepath", Type: cty.String},
		},
		VarParam: nil,
		Type:     function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.StringVal(filepath.Base(args[0].AsString())), nil
		},
	}),

	"path_join": function.New(&function.Spec{
		Description: "Join multiple paths elements into path string",
		VarParam: &function.Parameter{
			Name:             "elem",
			Description:      "Path elements",
			Type:             cty.String,
			AllowNull:        false,
			AllowUnknown:     false,
			AllowDynamicType: false,
			AllowMarked:      false,
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			elems := make([]string, len(args))
			for i := range args {
				elems[i] = args[i].AsString()
			}

			return cty.StringVal(filepath.Join(elems...)), nil
		},
	}),
}
