package manifest

import (
	"github.com/zclconf/go-cty/cty"
	"time"
)

// Attributes is user defined attributes
type Attributes = map[string]cty.Value

// JobType is job type
type JobType uint

const (
	// InvalidJob is invalid job type
	InvalidJob JobType = iota

	// ActionJob is job that calls action
	ActionJob

	// TaskJob is job that call another task
	TaskJob

	// MixinJob is job that calls mixin
	MixinJob
)

// Job is task or mixin step
type Job struct {
	// Type is job type. See JobType for more info.
	Type JobType

	// Target is name of action, task or mixin
	// that should be called
	Target string

	// Description is job description. optional
	Description string

	// Skip determines should job be skipped
	Skip bool

	// Async determines should job by async
	Async bool

	// Delay is time delay before job starts
	Delay time.Duration

	// Deadline is job execution deadline
	Deadline time.Duration

	// Context contains all information about accessible
	// values and functions for job
	Context Context
}

// GetAttributes returns job attributes
// that will be passed to mixin, action or task
func (j Job) GetAttributes() Attributes {
	return j.Context.ctx.Variables
}
