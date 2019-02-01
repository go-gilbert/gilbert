package scope

import (
	"os"
	"path/filepath"
)

type Vars map[string]string

type Context struct {
	Globals     Vars
	Variables   Vars
	processor   ExpressionProcessor
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
		processor: ExpressionProcessor{ctx: c},
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

func (c *Context) ExpandVariables(str string) (out string, err error) {
	return c.processor.ReadString(str)
}

// Environ gets list of OS environment variables with globals
func (c *Context) Environ() (env []string) {
	env = os.Environ()
	for k, v := range c.Globals {
		env = append(env, k+"="+v)
	}

	return
}
