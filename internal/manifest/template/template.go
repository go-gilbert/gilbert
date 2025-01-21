package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/go-gilbert/gilbert/internal/support/shell"
)

const (
	templateName   = "manifest"
	leftDelimiter  = "{{"
	rightDelimiter = "}}"
)

var (
	tplParser *template.Template
	functions = template.FuncMap{
		"slice": createSliceOperator,
		"yaml":  convertToYamlOperator,
		"shell": evalShellOperator,
		"split": splitStringOperator,
	}
)

func init() {
	tplParser = template.New(templateName).
		Funcs(functions).
		Delims(leftDelimiter, rightDelimiter)
}

func splitStringOperator(delimiter string, str string) []string {
	return strings.Split(strings.TrimSpace(str), delimiter)
}

func createSliceOperator(args ...any) []any {
	return args
}

func convertToYamlOperator(arg any) any {
	data, err := json.Marshal(arg)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func evalShellOperator(cmd string) string {
	proc := shell.PrepareCommand(cmd)
	data, err := proc.CombinedOutput()
	if err != nil {
		msg := fmt.Sprintf(`command %q returned error (%s)`, proc.Args, err)
		if data != nil {
			msg += "\n\n" + string(data)
		}

		panic(msg)
	}

	return string(data)
}

// CompileManifest compiles manifest from static go template
func CompileManifest(data []byte) ([]byte, error) {
	tpl, err := tplParser.Parse(string(data))
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := tpl.ExecuteTemplate(&out, templateName, nil); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
