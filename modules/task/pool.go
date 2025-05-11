package task

import (
	"time"

	"github.com/ethanrous/weblens/modules/errors"
)

var ErrChildTaskFailed = errors.New("child task failed")

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

type PoolCleanupFunc func(Pool)

type Pool interface {
	QueueTask(task Task) (err error)
	IsGlobal() bool
	GetRootPool() Pool
	GetWorkerPool() WorkerPool
	CreatedInTask() Task
	IncTaskCount(int)
	IncCompletedTasks(int)
	RemoveTask(string)
	AddError(Task)
	IsRoot() bool
	LockExit()
	UnlockExit()
	GetTotalTaskCount() int
	GetCompletedTaskCount() int
	HandleTaskExit(replacement bool) bool
	SignalAllQueued()
	Wait(bool, ...Task)
	Errors() []Task
	Status() PoolStatus
	AddCleanup(PoolCleanupFunc)
}

type WorkerPool interface {
	AddHit(time.Time, Task)
	NewTaskPool(bool, Task) Pool
}
