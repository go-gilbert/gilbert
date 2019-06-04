package manifest

import "github.com/go-gilbert/gilbert-sdk"

// TaskSet is a set of tasks declared in a manifest file
type TaskSet map[string]Task

// Task is a group of jobs
type Task []sdk.Job

// AsyncJobsCount returns count of async jobs in the task
func (t Task) AsyncJobsCount() (count int) {
	for i := range t {
		if t[i].Async {
			count++
		}
	}

	return count
}

// Clone creates a new task copy
func (t Task) Clone() Task {
	out := make(Task, 0, len(t))
	copy(out, t)
	return out
}
