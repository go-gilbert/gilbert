package scope

type Role uint

const (
	RoleRoot Role = iota
	RoleWorkflow
	RoleTask
	RoleMixin
	RoleJob
)

const (
	projectFieldsCount = 2
	scopeFieldsCount   = 2
)

const (
	InputsKey  = "inputs"
	ProjectKey = "project"
	EnvKey     = "env"
)

type Globals struct {
	Project ProjectInfo
	Env     map[string]string
}

// Scope holds all variables and inputs related to a workflow.
// Each scope references a parent and form a prototype chain.
type Scope struct {
	Root    *Scope
	Parent  *Scope
	Role    Role
	Globals Globals
	Inputs  map[string]any
	Consts  map[string]any
}

// ValueByName is a stub for expr.EvalContext compatibility.
//
// This method will be removed.
func (s *Scope) ValueByName(_ string) (string, bool) {
	return "", false
}

// Fork returns a new child scope referencing a parent scope.
func (s *Scope) Fork(newRole Role) *Scope {
	root := s.Root
	if root == nil {
		root = s
	}

	newScope := &Scope{
		Root:    root,
		Parent:  s,
		Role:    newRole,
		Globals: s.Globals,
		Inputs:  s.Inputs,
		Consts:  s.Consts,
	}

	// TODO: decide whether inherit inputs based on a role.

	return newScope
}

func (s *Scope) Values() map[string]any {
	// TODO: include parent scope.
	count := len(s.Consts) + scopeFieldsCount + projectFieldsCount
	m := make(map[string]any, count)
	for k, v := range s.Consts {
		m[k] = v
	}

	m[InputsKey] = s.Inputs
	m[ProjectKey] = s.Globals.Project.Values()
	m[EnvKey] = s.Globals.Env
	return m
}
