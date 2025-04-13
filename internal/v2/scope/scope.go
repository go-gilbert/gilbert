package scope

import (
	"iter"
	"maps"
)

type Role uint

const (
	RoleRoot Role = iota
	RoleWorkflow
	RoleTask
	RoleMixin
	RoleJob
	RoleClosure
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

	chain scopeChain
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

	newTail := &chainNode{
		value: s,
		prev:  s.chain.tail,
	}
	newChain := s.chain
	newChain.tail = newTail
	if newChain.head == nil {
		newChain.head = newTail
	}

	newScope := &Scope{
		Root:    root,
		Parent:  s,
		Role:    newRole,
		Globals: s.Globals,
		Consts:  s.Consts,
		Inputs:  map[string]any{}, // no need to copy as we've a reference to a parent.
		chain:   newChain,
	}

	return newScope
}

// Values exports scope values environment with visible variables for executing expressions.
//
// Exported values are inherited from parent scopes.
func (s *Scope) Values() (map[string]any, error) {
	// TODO: check if storing chain leaks memory
	globalsCount := len(s.Consts) + scopeFieldsCount + projectFieldsCount

	inputs := make(map[string]any, s.chain.inputsCount+len(s.Inputs))
	globals := make(map[string]any, s.chain.constCount+globalsCount)

	for parent := range iterChain(s.chain) {
		maps.Copy(inputs, parent.Inputs)
		maps.Copy(globals, parent.Consts)
	}

	maps.Copy(inputs, s.Inputs)
	maps.Copy(globals, s.Consts)

	globals[InputsKey] = inputs
	globals[ProjectKey] = s.Globals.Project.Values()
	globals[EnvKey] = s.Globals.Env
	return globals, nil
}

// Dispose detaches scope.
func (s *Scope) Dispose() {
	*s = Scope{}
}

type chainNode struct {
	value *Scope
	next  *chainNode
	prev  *chainNode
}

type scopeChain struct {
	inputsCount int
	constCount  int
	chainSize   int
	head        *chainNode
	tail        *chainNode
}

func iterChain(s scopeChain) iter.Seq[*Scope] {
	return func(yield func(*Scope) bool) {
		current := s.head
		for current != nil {
			if !yield(current.value) {
				return
			}
			current = current.next
		}
	}
}
