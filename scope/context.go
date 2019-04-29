package scope

import (
	sdk "github.com/go-gilbert/gilbert-sdk"
	"os"
	"path/filepath"
)

// Scope contains a set of globals and variables related to specific job
type Scope struct {
	Globals     sdk.Vars // Globals is set of global variables for all tasks
	Variables   sdk.Vars // Variables is set of variables for specific job
	processor   ExpressionProcessor
	environment sdk.ProjectEnvironment
}

// CreateScope creates a new context
func CreateScope(projectDirectory string, vars sdk.Vars) (c *Scope) {
	c = &Scope{
		Globals: sdk.Vars{
			"PROJECT": projectDirectory,
			"BUILD":   filepath.Join(projectDirectory, "build"),
			"GOPATH":  os.Getenv("GOPATH"),
		},
		Variables: vars,
	}

	c.processor = NewExpressionProcessor(c)
	c.environment.ProjectDirectory = projectDirectory
	return
}

func (c *Scope) Vars() sdk.Vars {
	return c.Variables
}

// Environment returns information about project environment
func (c *Scope) Environment() sdk.ProjectEnvironment {
	return c.environment
}

// AppendGlobals appends global variables to the context
func (c *Scope) AppendGlobals(globals sdk.Vars) sdk.ScopeAccessor {
	c.Globals = c.Globals.Append(globals)
	return c
}

// AppendVariables appends local variables to the context
func (c *Scope) AppendVariables(vars sdk.Vars) sdk.ScopeAccessor {
	c.Variables = c.Variables.Append(vars)
	return c
}

// Global returns a global variable value by it's name
func (c *Scope) Global(varName string) (out string, ok bool) {
	out, ok = c.Globals[varName]
	return
}

// Var returns a local variable value by it's name
func (c *Scope) Var(varName string) (isLocal bool, out string, ok bool) {
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
func (c *Scope) ExpandVariables(str string) (out string, err error) {
	return c.processor.ReadString(str)
}

// Scan does the same as ExpandVariables but with multiple variables and updates the value in pointer with expanded value
//
// Useful for bulk mapping of struct fields
func (c *Scope) Scan(vals ...*string) (err error) {
	for _, ptr := range vals {
		*ptr, err = c.processor.ReadString(*ptr)
		if err != nil {
			return err
		}
	}

	return nil
}

// Environ gets list of OS environment variables with globals
func (c *Scope) Environ() (env []string) {
	env = os.Environ()
	for k, v := range c.Globals {
		env = append(env, k+"="+v)
	}

	return
}
