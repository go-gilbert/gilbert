package env

import (
	"fmt"
	"os"
	"path/filepath"
)

const ExpPrefix = '$'
const ExpStart = '('
const ExpEnd = ')'

type Vars map[string]string

type Context struct {
	Globals     Vars
	Variables   Vars
	Environment struct {
		ProjectDirectory string
	}
}

func CreateContext(projectDirectory string, vars Vars) (c *Context) {
	c = &Context{
		Globals: Vars{
			"PROJECT": projectDirectory,
			"BUILD":   filepath.Join(projectDirectory, "build"),
			"GOPATH":  os.Getenv("GOPATH"),
		},
		Variables: vars,
	}

	c.Environment.ProjectDirectory = projectDirectory
	return
}

func (c *Context) AppendGlobals(globals Vars) *Context {
	for k, v := range globals {
		c.Globals[k] = v
	}

	return c
}

func (c *Context) Global(varName string) (out string, ok bool) {
	out, ok = c.Globals[varName]
	return
}

func (c *Context) Var(varName string) (isLocal bool, out string, ok bool) {
	out, ok = c.Variables[varName]
	if ok {
		isLocal = true
	}

	if !ok {
		out, ok = c.Globals[varName]
	}

	return
}

func (c *Context) isExpression(curPos int, str string) bool {
	strLen := len(str)
	if str[curPos] != ExpPrefix {
		return false
	}

	expStartPos := curPos + 1
	if strLen <= expStartPos || str[expStartPos] != ExpStart {
		return false
	}

	valueStartPos := expStartPos + 1
	if valueStartPos >= strLen {
		return false
	}

	return true
}

func (c *Context) ExpandVariables(str string) (out string, err error) {
	chars := make([]byte, 0)
	charCount := len(str)

	defer func() {
		if err == nil {
			out = string(chars)
		}
	}()

	for pos := 0; pos < charCount; {
		char := str[pos]
		if c.isExpression(pos, str) {
			parsed, end, eerr := c.parseStringSeqment(pos, str)
			if eerr != nil {
				err = eerr
				return
			}

			chars = append(chars, []byte(parsed)...)
			if end >= charCount {
				return
			}
			pos = end
			continue
		}

		chars = append(chars, char)
		pos++
	}

	return
}

func (c *Context) parseStringSeqment(pos int, str string) (out []byte, end int, err error) {
	start := pos + 2
	varName, end, err := c.expandExpression(start, str)
	if err != nil {
		return
	}

	isLocal, value, ok := c.Var(varName)
	if !ok {
		err = fmt.Errorf("variable '%s' is undefined", varName)
		return
	}

	// Parse local variable for nested global variables
	if isLocal {
		value, err = c.ExpandVariables(value)
		if err != nil {
			return nil, 0, err
		}
	}

	out = []byte(value)
	return
}

func (c *Context) expandExpression(position int, str string) (out string, end int, err error) {
	chars := make([]byte, 0)
	strLen := len(str)
	for pos := position; pos < strLen; pos++ {
		chr := str[pos]
		if chr == ExpEnd {
			end = pos + 1
			out = string(chars)
			return
		}

		chars = append(chars, chr)
	}

	err = fmt.Errorf("syntax error - expression is not finished at position %d: %s", position, str)
	return
}

// Environ gets list of OS environment variables with globals
func (c *Context) Environ() (env []string) {
	env = os.Environ()
	for k, v := range c.Globals {
		env = append(env, k+"="+v)
	}

	return
}
