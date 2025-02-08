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

type Job struct {
	Kind      JobKind
	Name      string
	Async     bool
	Delay     time.Duration
	Deadline  time.Duration
	Condition *LazyValue
	Inputs    map[string]*LazyValue
	Location  ReferenceLocation
}

type Task struct {
	Name        string
	Description string
	Inputs      map[string]InputSchema
	Jobs        []Job
	Location    ReferenceLocation
}

type JobFile struct {
	Vars   map[string]any
	Inputs map[string]InputSchema
	Tasks  map[string]Task
}
