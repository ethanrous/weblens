package task

type CleanupFunc func(Task)

type Task interface {
	Id() string
	Wait()
	Status() (bool, TaskExitStatus)
	SetResult(result TaskResult)
	GetResult() TaskResult
	SetCleanup(fn CleanupFunc)
	GetMeta() TaskMetadata
	GetTaskPool() Pool
	JobName() string
	SetChildTaskPool(pool Pool)
	ReadError() error
}
