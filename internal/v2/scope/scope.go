package scope

type Context uint

const (
	ContextRoot Context = iota
	ContextTask
	ContextMixin
	ContextJob
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

type Scope struct {
	Root    *Scope
	Parent  *Scope
	Context Context
	Globals Globals
	Inputs  map[string]any
	Consts  map[string]any
}

func (s *Scope) Values() map[string]any {
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
