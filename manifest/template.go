package manifest

import (
	"bytes"
	"encoding/json"
	"text/template"
)

const (
	templateName   = "manifest"
	leftDelimiter  = "{{{"
	rightDelimiter = "}}}"
)

var (
	tplParser *template.Template
	functions = template.FuncMap{
		"sliceOf": func(args ...interface{}) []interface{} {
			return args
		},
		"yaml": func(arg interface{}) interface{} {
			data, _ := json.Marshal(arg)
			return string(data)
		},
	}
)

func init() {
	tplParser = template.New(templateName).
		Funcs(functions).
		Delims(leftDelimiter, rightDelimiter)
}

func parseManifestTemplate(data []byte) ([]byte, error) {
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
