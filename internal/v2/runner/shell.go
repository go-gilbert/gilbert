package runner

type Notifier interface {
	NotifyTaskStarted(taskName string)
}
