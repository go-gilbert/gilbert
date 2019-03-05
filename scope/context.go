package scope

import (
	"os"
	"path/filepath"
)

// Vars is a set of declared variables
type Vars map[string]string

// Append appends variables from vars list
func (v Vars) Append(newVars Vars) {
	if newVars == nil || len(newVars) == 0 {
		return
	}

	for k := range newVars {
		v[k] = newVars[k]
	}
}

// Context contains a set of globals and variables related to specific job
type Context struct {
	Globals     Vars // Globals is set of global variables for all tasks
	Variables   Vars // Variables is set of variables for specific job
	processor   ExpressionProcessor
	Environment struct {
		ProjectDirectory string
	}
}

// CreateContext creates a new context
func CreateContext(projectDirectory string, vars Vars) (c *Context) {
	c = &Context{
		Globals: Vars{
			"PROJECT": projectDirectory,
			"BUILD":   filepath.Join(projectDirectory, "build"),
			"GOPATH":  os.Getenv("GOPATH"),
		},
		Variables: vars,
	}

	c.processor = NewExpressionProcessor(c)
	c.Environment.ProjectDirectory = projectDirectory
	return
}

// AppendGlobals appends global variables to the context
func (c *Context) AppendGlobals(globals Vars) *Context {
	for k, v := range globals {
		c.Globals[k] = v
	}

	return c
}

// AppendGlobals appends global variables to the context
func (c *Context) AppendVariables(globals Vars) *Context {
	for k, v := range globals {
		c.Globals[k] = v
	}

	return c
}

// Global returns a global variable value by it's name
func (c *Context) Global(varName string) (out string, ok bool) {
	out, ok = c.Globals[varName]
	return
}

// Var returns a local variable value by it's name
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

// ExpandVariables expands an expression stored inside a passed string
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
