package manifest

// Mixins is a pair of mixin name and params
type Mixins map[string]Mixin

// Mixin represents a mixin declaration
type Mixin []Job

// ToTask creates a new task from mixin with variables for override
func (m Mixin) ToTask(parentVars Vars) (t Task) {
	t = make(Task, 0, len(m))
	for _, j := range m {
		j.Vars = j.Vars.Append(parentVars)
		t = append(t, j)
	}

	return t
}
