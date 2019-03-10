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
