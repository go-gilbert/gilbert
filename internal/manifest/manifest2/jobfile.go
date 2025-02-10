package manifest2

import (
	"time"
)

type JobKind uint8

const (
	JobKindUnknown JobKind = iota
	JobKindAction
	JobKindMixin
	JobKindTask
)

type ExecStrategy struct {
	Delay   time.Duration
	Timeout time.Duration
	Matrix  map[string]*LazyValue
}

type Job struct {
	Location  ReferenceLocation
	Kind      JobKind
	Name      string
	Async     bool
	Strategy  ExecStrategy
	Condition *LazyValue
	Inputs    map[string]*LazyValue
}

type Task struct {
	Location    ReferenceLocation
	Name        string
	Description string
	Inputs      map[string]InputDefinition
	Jobs        []Job
}

type Mixin struct {
	Location    ReferenceLocation
	Name        string
	Description string
	Inputs      map[string]InputDefinition
	Jobs        []Job
}

type JobFile struct {
	Consts map[string]any
	Inputs map[string]InputDefinition
	Tasks  map[string]*Task
	Mixins map[string]*Mixin
}
