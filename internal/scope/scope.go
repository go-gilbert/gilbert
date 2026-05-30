package scope

import "path/filepath"

const (
	projectFieldsCount = 2
	scopeFieldsCount   = 2
)

const (
	InputsKey  = "inputs"
	ProjectKey = "project"
	EnvKey     = "env"
)

type ProjectInfo struct {
	WorkDir      string `expr:"workDir"`
	WorkspaceDir string `expr:"workspaceDir"`
	WorkflowFile string `expr:"workflowFile"`
}

type Globals struct {
	Project ProjectInfo
	Env     map[string]string
}

// Scope holds all variables and inputs related to a workflow.
// Each scope references a parent and form a prototype chain.
type Scope struct {
	// Root points to a root scope.
	Root *Scope

	// Parent points to a parent scope.
	Parent *Scope

	// Globals is global execution information (environment variables, project, etc).
	Globals Globals

	// Environment holds scope-local environment variables.
	Environment map[string]string

	// Inputs is job parameter values.
	Inputs map[string]any

	// Consts is constant values declared in a workflow.
	Consts map[string]any

	// MatrixValues is values populated by job matrix execution strategy.
	MatrixValues map[string]any

	// EventData contains signal event data.
	//
	// Populated to jobs called by a signal.
	// Contents of event data depends on event and sender.
	EventData map[string]any
}

// ValueByName is a stub for expr.EvalContext compatibility.
//
// This method will be removed.
func (s *Scope) ValueByName(_ string) (string, bool) {
	return "", false
}

// Fork returns a new child scope referencing a parent scope.
func (s *Scope) Fork() *Scope {
	root := s.Root
	if root == nil {
		root = s
	}

	newScope := &Scope{
		Root:         root,
		Parent:       s,
		Globals:      s.Globals,
		Consts:       s.Consts,
		Inputs:       map[string]any{}, // no need to copy as we've a reference to a parent.
		MatrixValues: s.MatrixValues,
		EventData:    s.EventData,
	}

	return newScope
}

// GetEnv searches and returns environment variable value by traversing scope chain and globals.
func (s *Scope) GetEnv(key string) (string, bool) {
	// Environment variables can be overriden per scope.
	// First, traverse a chain and use globals only as a fallback.
	sk := s
	for {
		if sk == nil {
			break
		}

		env := sk.Environment
		if len(env) > 0 {
			v, ok := env[key]
			if ok {
				return v, true
			}
		}

		sk = sk.Parent
	}

	v, ok := s.Globals.Env[key]
	return v, ok
}

// WithMatrixValues sets job matrix values, overwriting existing value.
func (s *Scope) WithMatrixValues(labels []string, values []any) *Scope {
	m := make(map[string]any, len(labels))
	for i, k := range labels {
		if i < len(values) {
			m[k] = values[i]
		}
	}

	s.MatrixValues = m
	return s
}

// WithWorkDir updates working directory of current scope.
func (s *Scope) WithWorkDir(wd string) *Scope {
	if wd == "" {
		return s
	}

	newWd := wd
	if !filepath.IsAbs(newWd) {
		newWd = filepath.Join(s.Globals.Project.WorkDir, newWd)
	}

	s.Globals.Project.WorkDir = filepath.Clean(newWd)
	return s
}

// Values exports scope values environment with visible variables for executing expressions.
//
// Exported values are inherited from parent scopes.
func (s *Scope) Values() (any, error) {
	return newExprEnvironment(s), nil
}

// Dispose detaches scope.
func (s *Scope) Dispose() {
	*s = Scope{}
}
