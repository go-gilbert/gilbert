package manifest

import (
	"github.com/go-gilbert/gilbert-sdk"
)

// Mixin represents a mixin declaration
type Mixin []sdk.Job

// ToTask creates a new task from mixin with variables for override
func (m Mixin) ToTask(parentVars sdk.Vars) (t Task) {
	t = make(Task, 0, len(m))
	for _, j := range m {
		j.Vars = j.Vars.Append(parentVars)
		t = append(t, j)
	}

	return t
}

// Mixins is a pair of mixin name and params
type Mixins map[string]Mixin
