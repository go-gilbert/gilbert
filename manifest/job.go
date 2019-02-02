package manifest

import "github.com/x1unix/gilbert/scope"

// Job is a single job in task
type Job struct {
	// Condition is shell command that should be successful to run specified job
	Condition string `yaml:"if,omitempty"`

	// Description is job description
	Description string `yaml:"description,omitempty"`

	// Task refers to task that should be run.
	Task string `yaml:"run,omitempty"`

	// Plugin describes what plugin should handle this job.
	Plugin string `yaml:"plugin,omitempty"`

	// Vars is a set of variables defined for this job.
	Vars scope.Vars `yaml:"vars,omitempty"`

	// Params is a set of arguments for the job.
	Params map[string]interface{} `yaml:"params,omitempty"`
}

// InvokesTask checks if this job should call another task
func (j *Job) InvokesTask() bool {
	return j.Task != ""
}

// InvokesPlugin checks if this job should invoke plugin
func (j *Job) InvokesPlugin() bool {
	return j.Plugin != ""
}

// HasDescription checks if description is available
func (j *Job) HasDescription() bool {
	return j.Description != ""
}
