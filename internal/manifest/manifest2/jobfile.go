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

// DocHeader is generic structure for embedding field metadata such as name, location or doc string.
type DocHeader struct {
	Name string
	Doc  []string
}

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
	DocHeader
	Position ReferenceLocation
	Inputs   map[string]InputDefinition
	Jobs     []Job
}

type JobFile struct {
	Consts map[string]any
	Inputs map[string]*InputDefinition
	Tasks  map[string]*Task
	Mixins map[string]*Task
}
