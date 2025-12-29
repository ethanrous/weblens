package task

import (
	"time"

	"github.com/ethanrous/weblens/modules/errors"
)

// ErrChildTaskFailed indicates that one or more child tasks in a pool failed.
var ErrChildTaskFailed = errors.New("child task failed")

// PoolStatus represents the current state and progress of a task pool.
type PoolStatus struct {
	// The count of tasks that have completed on this task pool.
	// Complete *DOES* include failed tasks
	Complete int64

	// The count of failed tasks on this task pool
	Failed int

	// The count of all tasks that have been queued on this task pool
	Total int64

	// Percent to completion of all tasks
	Progress float64

	// How long the pool has been alive
	Runtime time.Duration
}

// PoolCleanupFunc is a function that performs cleanup operations when a pool completes.
type PoolCleanupFunc func(Pool)

// Pool manages the execution and lifecycle of a group of tasks.
type Pool interface {
	QueueTask(task Task) (err error)
	IsGlobal() bool
	GetRootPool() Pool
	GetWorkerPool() WorkerPool
	CreatedInTask() Task
	IncTaskCount(incCount int)
	IncCompletedTasks(incCount int)
	RemoveTask(taskID string)
	AddError(task Task)
	IsRoot() bool
	LockExit()
	UnlockExit()
	GetTotalTaskCount() int
	GetCompletedTaskCount() int
	HandleTaskExit(replacement bool) bool
	SignalAllQueued()
	Wait(supplementWorker bool, activeTask ...Task)
	Errors() []Task
	Status() PoolStatus
	AddCleanup(fn PoolCleanupFunc)
}

// WorkerPool manages worker threads and creates task pools for job execution.
type WorkerPool interface {
	AddHit(hitTime time.Time, task Task)
	NewTaskPool(replace bool, activeTask Task) Pool
}
