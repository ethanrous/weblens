package types

import "time"

type Task interface {
	TaskId() TaskId
	TaskType() TaskType
	Status() (bool, TaskExitStatus)
	GetResult(...string) map[string]any
	Q(TaskPool) Task
	Wait()
	Cancel()
	SwLap(string)
	// SetCaster(BroadcasterAgent)
	ClearTimeout()

	ReadError() any
	ClearAndRecompute()

	SetErrorCleanup(cleanup func())
	SetCleanup(cleanup func())

	AddChunkToStream(FileId, []byte, string) error
	NewFileInStream(WeblensFile, int64) error
	ExeTime() time.Duration
}

// Tasker interface for queueing tasks in the task pool
type TaskPool interface {
	MarkGlobal()
	QueueTask(Task) error
	SignalAllQueued()
	Wait(bool)
	NotifyTaskComplete(Task, BroadcasterAgent, ...any)

	ScanDirectory(WeblensFile, bool, bool, BroadcasterAgent) Task
	ScanFile(file WeblensFile, broadcaster BroadcasterAgent) Task
	WriteToFile(FileId, int64, int64, BroadcasterAgent) Task
	MoveFile(FileId, FileId, string, BroadcasterAgent) Task
	GatherFsStats(WeblensFile, BroadcasterAgent) Task
	Backup(string, Requester) Task
	HashFile(file WeblensFile) Task

	Errors() []Task
	AddError(t Task)
}

type TaskId string

func (tId TaskId) String() string {
	return string(tId)
}

type TaskType string
type TaskExitStatus string
type TaskEvent string
type TaskResult map[string]any
