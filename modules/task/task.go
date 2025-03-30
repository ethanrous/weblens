package task

type Task interface {
	Id() string
	Wait()
	Status() (bool, TaskExitStatus)
	GetResult() TaskResult
}
