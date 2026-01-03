package task

// ExitStatus represents the final state of a completed task.
type ExitStatus string

const (
	// TaskNoStatus indicates the task has not yet completed.
	TaskNoStatus ExitStatus = ""
	// TaskSuccess indicates the task completed successfully.
	TaskSuccess ExitStatus = "success"
	// TaskCanceled indicates the task was canceled before completion.
	TaskCanceled ExitStatus = "canceled"
	// TaskError indicates the task failed with an error.
	TaskError ExitStatus = "error"
)
