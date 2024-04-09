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

	ReadError() any
	ClearAndRecompute()

	AddChunkToStream(FileId, []byte, string) error
	ExeTime() time.Duration
}

// Tasker interface for queueing tasks in the task pool
type TaskPool interface {
	MarkGlobal()
	QueueTask(Task) error
	SignalAllQueued()
	Wait(bool)

	ScanDirectory(WeblensFile, bool, bool, BroadcasterAgent) Task
	ScanFile(WeblensFile, Media, BroadcasterAgent) Task
	WriteToFile(FileId, int64, int64, BroadcasterAgent) Task
	MoveFile(FileId, FileId, string, BroadcasterAgent) Task
	GatherFsStats(WeblensFile, BroadcasterAgent) Task
	Backup() Task
}

type TaskId string

func (tId TaskId) String() string {
	return string(tId)
}

type TaskType string
type TaskExitStatus string
type TaskResult map[string]any
