package scope

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

	// Inputs is job parameter values.
	Inputs map[string]any

	// Consts is constant values declared in a workflow.
	Consts map[string]any

	// MatrixValues is values populated by job matrix execution strategy.
	MatrixValues map[string]any
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
		Root:    root,
		Parent:  s,
		Globals: s.Globals,
		Consts:  s.Consts,
		Inputs:  map[string]any{}, // no need to copy as we've a reference to a parent.
	}

	return newScope
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
