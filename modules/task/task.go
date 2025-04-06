package task

type CleanupFunc func(Task)

type Task interface {
	Id() string
	Wait()
	Status() (bool, TaskExitStatus)
	GetResult() TaskResult
	SetCleanup(CleanupFunc)
	GetMeta() TaskMetadata
	GetTaskPool() Pool
	JobName() string
	SetChildTaskPool(Pool)
	ReadError() error
}
