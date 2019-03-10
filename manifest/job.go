package manifest

import (
	"github.com/x1unix/gilbert/scope"
	"time"
)

// JobExecType represents job type
type JobExecType uint8

const (
	// ExecEmpty means job has no execution type
	ExecEmpty JobExecType = iota

	// ExecPlugin means that job execute plugin
	ExecPlugin

	// ExecTask means that job runs other task
	ExecTask

	// ExecMixin means that job based on mixin
	ExecMixin
)

// Period is job period in milliseconds
type Period uint

// ToDuration returns value in milliseconds for time.Duration
func (d Period) ToDuration() time.Duration {
	return time.Duration(d) * time.Millisecond
}

// Job is a single job in task
type Job struct {
	// Condition is shell command that should be successful to run specified job
	Condition string `yaml:"if,omitempty" mapstructure:"if"`

	// Description is job description
	Description string `yaml:"description,omitempty" mapstructure:"description"`

	// TaskName refers to task that should be run.
	TaskName string `yaml:"run,omitempty" mapstructure:"run"`

	// PluginName describes what plugin should handle this job.
	PluginName string `yaml:"plugin,omitempty" mapstructure:"plugin"`

	// MixinName is mixin to be used by this job
	MixinName string `yaml:"mixin,omitempty" mapstructure:"mixin"`

	// Async means that job should be run asynchronously
	Async bool `yaml:"async,omitempty" mapstructure:"async"`

	// Delay before task start in milliseconds
	Delay Period `yaml:"delay,omitempty" mapstructure:"delay"`

	// Period is a time quota for job
	Deadline Period `yaml:"deadline,omitempty" mapstructure:"deadline"`

	// Vars is a set of variables defined for this job.
	Vars scope.Vars `yaml:"vars,omitempty" mapstructure:"vars"`

	// Params is a set of arguments for the job.
	Params map[string]interface{} `yaml:"params,omitempty" mapstructure:"params"`
}

// HasDescription checks if description is available
func (j *Job) HasDescription() bool {
	return j.Description != ""
}

// FormatDescription returns formatted description string
func (j *Job) FormatDescription() string {
	if j.Description != "" {
		return j.Description
	}

	// If description is empty, return used mixin or plugin name if available
	for _, v := range []string{j.PluginName, j.TaskName, j.MixinName} {
		if v != "" {
			return v
		}
	}

	return ""
}

// Type returns job execution type
//
// If job has no 'plugin', 'task' or 'plugin' declaration, ExecEmpty will be returned
func (j *Job) Type() JobExecType {
	if j.PluginName != "" {
		return ExecPlugin
	}

	if j.TaskName != "" {
		return ExecTask
	}

	if j.MixinName != "" {
		return ExecMixin
	}

	return ExecEmpty
}
