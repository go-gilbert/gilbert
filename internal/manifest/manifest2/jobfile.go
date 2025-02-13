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

type JobGroupType uint8

const (
	JobGroupKindUnknown JobGroupType = iota
	JobGroupTypeTask
	JobGroupTypeMixin
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

// JobGroup is collection of jobs to execute with input parameters.
//
// Acts as a base for tasks and mixins.
type JobGroup struct {
	DocHeader
	Type     JobGroupType
	Position ReferenceLocation
	Inputs   map[string]InputDefinition
	Jobs     []Job
}

type JobFile struct {
	Consts map[string]any
	Inputs map[string]*InputDefinition
	Tasks  map[string]*JobGroup
	Mixins map[string]*JobGroup
}
