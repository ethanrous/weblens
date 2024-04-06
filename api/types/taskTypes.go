package types

import "time"

type Task interface {
	TaskId() TaskId
	TaskType() TaskType
	Status() (bool, TaskExitStatus)
	GetResult(...string) map[string]any
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
type TaskerAgent interface {

	// Parameters:
	//
	//	- `directory` : the weblens file descriptor representing the directory to scan
	//
	//	- `recursive` : if true, scan all children of directory recursively
	//
	//	- `deep` : query and sync with the real underlying filesystem for changes not reflected in the current fileTree
	ScanDirectory(WeblensFile, bool, bool, BroadcasterAgent) Task

	ScanFile(file WeblensFile, m Media, caster BroadcasterAgent) Task
	MarkGlobal()
	WriteToFile(FileId, int64, int64, BroadcasterAgent) Task
}

type TaskId string

func (tId TaskId) String() string {
	return string(tId)
}

type TaskType string
type TaskExitStatus string
type TaskResult map[string]any
