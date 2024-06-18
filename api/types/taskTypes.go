package types

import "time"

type Task interface {
	TaskId() TaskId
	TaskType() TaskType
	GetTaskPool() TaskPool
	Status() (bool, TaskExitStatus)
	GetResult(string) any
	GetResults() map[string]any

	Q(TaskPool) Task

	Wait() Task
	Cancel()

	SwLap(string)
	ClearTimeout()

	ReadError() any
	ClearAndRecompute()
	SetCleanup(cleanup func())
	SetErrorCleanup(cleanup func())

	AddChunkToStream(FileId, []byte, string) error
	NewFileInStream(WeblensFile, int64) error
	ExeTime() time.Duration
}

// TaskPool is the interface for grouping and sending tasks to be processed in a WorkerPool
type TaskPool interface {
	MarkGlobal()
	IsGlobal() bool
	QueueTask(Task) error
	SignalAllQueued()
	Wait(bool)
	NotifyTaskComplete(Task, BroadcasterAgent, ...any)
	CreatedInTask() Task
	IsRoot() bool
	GetWorkerPool() WorkerPool
	Status() (int, int, float64)

	GetRootPool() TaskPool

	LockExit()
	UnlockExit()

	ScanDirectory(WeblensFile, BroadcasterAgent) Task
	ScanFile(file WeblensFile, broadcaster BroadcasterAgent) Task
	WriteToFile(FileId, int64, int64, BroadcasterAgent) Task
	MoveFile(FileId, FileId, string, BroadcasterAgent) Task
	GatherFsStats(WeblensFile, BroadcasterAgent) Task
	Backup(InstanceId, Requester, FileTree) Task
	HashFile(WeblensFile, BroadcasterAgent) Task
	CreateZip(files []WeblensFile, username Username, shareId ShareId, ft FileTree, casters BroadcasterAgent) Task

	Errors() []Task
	AddError(t Task)
}

type WorkerPool interface {
	Run()
	NewTaskPool(replace bool, createdBy Task) TaskPool
	GetTask(taskId TaskId) Task
	AddHit(time time.Time, target Task)
}

type TaskId string

func (tId TaskId) String() string {
	return string(tId)
}

type TaskType string
type TaskExitStatus string
type TaskEvent string
type TaskResult map[string]any
