package types

import (
	"time"
)

type Task interface {
	TaskId() TaskId
	TaskType() TaskType
	GetTaskPool() TaskPool
	GetChildTaskPool() TaskPool
	Status() (bool, TaskExitStatus)
	GetMeta() TaskMetadata
	GetResult(string) any
	GetResults() TaskResult

	Q(TaskPool) Task

	Wait() Task
	Cancel()

	SwLap(string)
	ClearTimeout()

	ReadError() any
	ClearAndRecompute()
	SetPostAction(action func(TaskResult))
	SetCleanup(cleanup func())
	SetErrorCleanup(cleanup func())

	AddChunkToStream(FileId, []byte, string) error
	NewFileInStream(WeblensFile, int64) error
	ExeTime() time.Duration
}

// TaskPool is the interface for grouping and sending tasks to be processed in a WorkerPool
type TaskPool interface {
	ID() TaskId

	QueueTask(Task) error

	MarkGlobal()
	IsGlobal() bool
	SignalAllQueued()

	CreatedInTask() Task

	Wait(bool)
	Cancel()

	IsRoot() bool
	Status() TaskPoolStatus
	AddCleanup(fn func())

	GetRootPool() TaskPool
	GetWorkerPool() WorkerPool

	LockExit()
	UnlockExit()

	RemoveTask(TaskId)

	NotifyTaskComplete(Task, BroadcasterAgent, ...any)

	ScanDirectory(WeblensFile, BroadcasterAgent) Task
	ScanFile(WeblensFile, BroadcasterAgent) Task
	WriteToFile(FileId, int64, int64, BroadcasterAgent) Task
	MoveFile(FileId, FileId, string, FileEvent, BroadcasterAgent) Task
	GatherFsStats(WeblensFile, BroadcasterAgent) Task
	Backup(InstanceId, BroadcasterAgent) Task
	HashFile(WeblensFile, BroadcasterAgent) Task
	CreateZip(files []WeblensFile, username Username, shareId ShareId, casters BroadcasterAgent) Task
	CopyFileFromCore(WeblensFile, BroadcasterAgent) Task

	Errors() []Task
	AddError(t Task)
}

type WorkerPool interface {
	Run()
	NewTaskPool(replace bool, createdBy Task) TaskPool
	GetTask(taskId TaskId) Task
	AddHit(time time.Time, target Task)
	GetTaskPool(TaskId) TaskPool
	GetTaskPoolByTaskType(taskType TaskType) TaskPool
}

type TaskId string

func (tId TaskId) String() string {
	return string(tId)
}

type TaskPoolStatus struct {
	// The count of tasks that have completed on this task pool
	Complete int64

	// The count of all tasks that have been queued on this task pool
	Total int64

	// Percent to completion of all tasks
	Progress float64

	// How long the pool has been alive
	Runtime time.Duration
}

type TaskType string
type TaskExitStatus string
type TaskEvent string
type TaskResult map[string]any

type TaskMetadata interface {
	MetaString() string
	FormatToResult() TaskResult
}
