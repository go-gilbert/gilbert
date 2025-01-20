package manifest

// TaskSet is a set of tasks declared in a manifest file
type TaskSet map[string]Task

// Task is a group of jobs
type Task []Job

// AsyncJobsCount returns count of async jobs in the task
func (t Task) AsyncJobsCount() (count int) {
	for i := range t {
		if t[i].Async {
			count++
		}
	}

	return count
}

// Clone creates a new task copy with specified variables
func (t Task) Clone(vars Vars) Task {
	out := make(Task, len(t))
	for i, j := range t {
		j.Vars = j.Vars.Append(vars)
		out[i] = j
	}

	return out
}
