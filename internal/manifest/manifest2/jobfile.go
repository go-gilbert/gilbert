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
	// Location contains information about where job is defined.
	Location ReferenceLocation

	// Kind defines whether job is a mixin or action call.
	Kind JobKind

	// Name is action or mixin name.
	Name string

	// Async tells whether runner should wait until job finishes before starting next job.
	Async bool

	// Strategy is job execution parameters.
	Strategy ExecStrategy

	// Condition is expression to check whether job should be executed.
	Condition *LazyValue

	// Args is job arguments.
	Args map[string]*LazyValue

	// Hooks is key-value pair of event name and actions to be run on event.
	Hooks map[string][]Job
}

type JobGroups = map[string]*JobGroup

// JobGroup is collection of jobs to execute with input parameters.
//
// Acts as a base for tasks and mixins.
type JobGroup struct {
	DocHeader
	Type     JobGroupType
	Location *ReferenceLocation
	Inputs   Inputs
	Jobs     []Job
}

type JobFile struct {
	Consts map[string]any
	Inputs Inputs
	Tasks  JobGroups
	Mixins JobGroups
}
