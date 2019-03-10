package scope

import (
	"os"
	"path/filepath"
)

// Vars is a set of declared variables
type Vars map[string]string

// Append appends variables from vars list
func (v Vars) Append(newVars Vars) (out Vars) {
	if v == nil {
		return newVars.Clone()
	}

	out = v.Clone()
	if newVars == nil || len(newVars) == 0 {
		return out
	}

	for k, val := range newVars {
		out[k] = val
	}

	return out
}

// Clone creates a copy of variables map
func (v Vars) Clone() (out Vars) {
	out = make(Vars, len(v))
	for k, val := range v {
		out[k] = val
	}

	return out
}

// Scope contains a set of globals and variables related to specific job
type Scope struct {
	Globals     Vars // Globals is set of global variables for all tasks
	Variables   Vars // Variables is set of variables for specific job
	processor   ExpressionProcessor
	Environment struct {
		ProjectDirectory string
	}
}

// CreateScope creates a new context
func CreateScope(projectDirectory string, vars Vars) (c *Scope) {
	c = &Scope{
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
func (c *Scope) AppendGlobals(globals Vars) *Scope {
	c.Globals = c.Globals.Append(globals)
	return c
}

// AppendVariables appends local variables to the context
func (c *Scope) AppendVariables(vars Vars) *Scope {
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
