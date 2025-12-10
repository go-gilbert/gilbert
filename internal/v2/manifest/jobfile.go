package manifest

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

// MatrixParam defines a matrix strategy parameter and values.
type MatrixParam struct {
	// Key is parameter name.
	Key string

	// Values is lazy-evaluated thunk that provides an array of parameter values.
	Values *LazyValue
}

type MatrixParamAndValues = map[string]any

// ExecStrategy defines matrix strategy for job execution.
//
// Matrix strategy lets to use variables in a single job definition to
// automatically create multiple job runs that are based on the combinations of the variables.
type ExecStrategy struct {
	// MaxParallel limits a number of concurrent jobs.
	//
	// Default: 1
	MaxParallel int

	// ContinueOnError determines whether matrix should continue running even if job fails.
	ContinueOnError bool

	// Matrix is ordered set of job configurations.
	Matrix []MatrixParam

	// MatrixKeys is set of field names used in [MatrixKeys].
	// Used to speed up key lookup and validation.
	//
	// Populated automatically by a parser from [MatrixKeys].
	MatrixKeys map[string]struct{}

	// Exclude is a list of configurations to exclude from running.
	//
	// Excluded configuration only has to be a partial match for it to be excluded.
	Exclude []MatrixParamAndValues
}

// SetMatrixParams sets [Matrix] and [MatrixKeys] values.
//
// Use this method instead of changing fields manually.
func (es *ExecStrategy) SetMatrixParams(params []MatrixParam) {
	m := make(map[string]struct{}, len(params))
	for _, v := range params {
		m[v.Key] = struct{}{}
	}

	es.Matrix = params
	es.MatrixKeys = m
}

type Job struct {
	// Location contains information about where job is defined.
	Location ReferenceLocation

	// Kind defines whether job is a mixin or action call.
	Kind JobKind

	// Handler specifies mixin or action to run a job.
	Handler JobHandlerRef

	// Async tells whether runner should wait until job finishes before starting next job.
	Async bool

	// Strategy sets up matrix execution strategy.
	Strategy ExecStrategy

	// Delay is interval to wait before executing a job.
	Delay time.Duration

	// Timeout is maximum job execution duration limit.
	Timeout time.Duration

	// Condition is expression to check whether job should be executed.
	//
	// When execution strategy is defined, block is applied to a whole job block.
	Condition *LazyValue

	// Args contains arguments passed to action or mixin.
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
	Path    string
	Plugins PluginImports
	Consts  map[string]any
	Inputs  Inputs
	Tasks   JobGroups
	Mixins  JobGroups
}
