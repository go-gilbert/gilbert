package manifest

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Context is task execution context that contains all values
// accessible to jobs and called mixins, sub-tasks and actions
//
// It contains user-defined job parameters or data returned
// from job execution
type Context struct {
	ctx *hcl.EvalContext
}

// JobsContainer is base structure for task and mixin structs.
//
// It contains name, description of task/mixin, required params
// and execution context.
type JobsContainer struct {
	blocks      hclsyntax.Blocks
	Name        string
	Description string
	Context     Context
	Parameters  Parameters
}

// Tasks is map of tasks
type Tasks map[string]Task

type Task struct {
	JobsContainer
}

func newTask(container JobsContainer) Task {
	return Task{
		JobsContainer: container,
	}
}
