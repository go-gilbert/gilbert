package scope

import (
	"fmt"

	exprRuntime "github.com/expr-lang/expr/vm/runtime"
)

var _ exprRuntime.Proxy = (*valueProxy)(nil)

type valueLookupFunc func(scope *Scope, key string) (any, bool)

// valueProxy is expr lang proxy helper to access values across a whole scope prototype chain.
type valueProxy struct {
	scope      *Scope
	lookupFunc valueLookupFunc
}

func newValueProxy(s *Scope, fn valueLookupFunc) *valueProxy {
	return &valueProxy{
		scope:      s,
		lookupFunc: fn,
	}
}

// GetProperty implements runtime.Proxy.
//
// Method performs lookup of a passed key across a whole scope inheritance tree.
func (s valueProxy) GetProperty(k any) any {
	key, ok := k.(string)
	if !ok {
		// expr lang threats panics as errors
		panic(fmt.Sprintf("key should be a string, got %T", k))
	}

	current := s.scope
	for {
		if current == nil {
			return nil
		}

		val, ok := s.lookupFunc(current, key)
		if ok {
			return val
		}

		current = current.Parent
	}
}

var (
	inputsLookupFunc valueLookupFunc = func(scope *Scope, key string) (any, bool) {
		v, ok := scope.Inputs[key]
		return v, ok
	}

	constsLookupFunc valueLookupFunc = func(scope *Scope, key string) (any, bool) {
		v, ok := scope.Consts[key]
		return v, ok
	}

	envLookupFunc valueLookupFunc = func(scope *Scope, key string) (any, bool) {
		// Env vars are overridable, thus needs to be proxied.
		if v, ok := scope.GetEnv(key); ok {
			return v, true
		}

		return nil, false
	}
)

// exprEnvironment is expr lang evaluation environment created from a scope.
type exprEnvironment struct {
	Inputs  *valueProxy    `expr:"inputs"`
	Consts  *valueProxy    `expr:"consts"`
	Env     *valueProxy    `expr:"env"`
	Project ProjectInfo    `expr:"project"`
	Matrix  map[string]any `expr:"matrix"`
	Event   map[string]any `expr:"event"`
}

func newExprEnvironment(s *Scope) *exprEnvironment {
	if s == nil {
		return &exprEnvironment{}
	}

	return &exprEnvironment{
		Inputs:  newValueProxy(s, inputsLookupFunc),
		Consts:  newValueProxy(s, constsLookupFunc),
		Env:     newValueProxy(s, envLookupFunc),
		Matrix:  s.MatrixValues,
		Project: s.Globals.Project,
		Event:   s.EventData,
	}
}
