package manifest

import "github.com/x1unix/guru/env"

// Job is a single job in task
type Job struct {
	// Description is job description
	Description string `yaml:"description"`

	// Task refers to task that should be run.
	Task string `yaml:"run"`

	// Plugin describes what plugin should handle this job.
	Plugin string `yaml:"plugin"`

	// Vars is a set of variables defined for this job.
	Vars env.Vars `yaml:"vars"`

	// Params is a set of arguments for the job.
	Params map[string]interface{} `yaml:"params"`
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
