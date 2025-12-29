package task

// CleanupFunc is a function called when a task completes.
type CleanupFunc func(Task)

// Task is an interface for background task execution.
type Task interface {
	ID() string
	Wait()
	Status() (bool, ExitStatus)
	SetResult(result Result)
	GetResult() Result
	SetCleanup(fn CleanupFunc)
	GetMeta() Metadata
	GetTaskPool() Pool
	JobName() string
	SetChildTaskPool(pool Pool)
	ReadError() error
}
