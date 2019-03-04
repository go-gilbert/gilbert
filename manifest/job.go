package manifest

import (
	"github.com/x1unix/gilbert/scope"
)

// Job is a single job in task
type Job struct {
	// Condition is shell command that should be successful to run specified job
	Condition string `yaml:"if,omitempty"`

	// Description is job description
	Description string `yaml:"description,omitempty"`

	// TaskName refers to task that should be run.
	TaskName *string `yaml:"run,omitempty"`

	// PluginName describes what plugin should handle this job.
	PluginName *string `yaml:"plugin,omitempty"`

	// MixinName is mixin to be used by this job
	MixinName *string `yaml:"mixin,omitempty"`

	// Delay before task start in milliseconds
	Delay uint `yaml:"delay,omitempty"`

	// Vars is a set of variables defined for this job.
	Vars scope.Vars `yaml:"vars,omitempty"`

	// Params is a set of arguments for the job.
	Params map[string]interface{} `yaml:"params,omitempty"`
}

// HasDescription checks if description is available
func (j *Job) HasDescription() bool {
	return j.Description != ""
}
