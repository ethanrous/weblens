package task

import (
	"time"

	task_mod "github.com/ethanrous/weblens/modules/task"
)

// NewTestTask creates a Task for testing purposes.
// This allows external packages to create Task instances with specific state.
// Only available when building with -tags=test.
func NewTestTask(id, jobName string, workerID int64, exitStatus task_mod.ExitStatus, result task_mod.Result, startTime time.Time) *Task {
	return &Task{
		taskID:     id,
		jobName:    jobName,
		WorkerID:   workerID,
		exitStatus: exitStatus,
		queueState: Exited,
		result:     result,
		StartTime:  startTime,
	}
}
