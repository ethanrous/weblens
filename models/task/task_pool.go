package task

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/rs/zerolog"
)

// GlobalTaskPoolID is the identifier for the global task pool that is always available.
const GlobalTaskPoolID = "GLOBAL"

// ErrChildTaskFailed indicates that one or more child tasks in a pool failed.
var ErrChildTaskFailed = wlerrors.New("child task failed")

// Pool manages a collection of tasks and tracks their execution progress.
type Pool struct {
	allQueuedFlag  atomic.Bool
	cleanupFn      []PoolCleanupFunc
	cleanupsDone   atomic.Bool
	completedTasks atomic.Int64
	createdAt      time.Time
	createdBy      *Task
	erroredTasks   []*Task
	exitLock       sync.Mutex
	hasQueueThread bool
	id             string
	log            zerolog.Logger
	parentTaskPool *Pool
	taskLock       sync.RWMutex
	tasks          map[string]*Task
	totalTasks     atomic.Int64
	treatAsGlobal  bool
	waiterCount    atomic.Int32
	waiterGate     chan struct{}
	workerPool     *WorkerPool
}

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

// PoolCleanupFunc defines a function type for cleaning up resources associated with a task pool.
type PoolCleanupFunc func(tp *Pool)

// IsRoot returns true if this task pool has no parent or its parent is global.
func (tp *Pool) IsRoot() bool {
	if tp == nil {
		return false
	}

	return tp.parentTaskPool == nil || tp.parentTaskPool.IsGlobal()
}

// GetWorkerPool returns the worker pool that manages this task pool.
func (tp *Pool) GetWorkerPool() *WorkerPool {
	return tp.workerPool
}

// ID returns the unique identifier for this task pool.
func (tp *Pool) ID() string {
	return tp.id
}

// IncTaskCount increments the total task count for this pool by the specified amount.
func (tp *Pool) IncTaskCount(count int) {
	tp.totalTasks.Add(int64(count))
}

// GetTotalTaskCount returns the total number of tasks that have been added to this pool.
func (tp *Pool) GetTotalTaskCount() int {
	return int(tp.totalTasks.Load())
}

// IncCompletedTasks increments the completed task count for this pool by the specified amount.
func (tp *Pool) IncCompletedTasks(count int) {
	tp.completedTasks.Add(int64(count))
}

// GetCompletedTaskCount returns the number of tasks that have completed in this pool.
func (tp *Pool) GetCompletedTaskCount() int {
	return int(tp.completedTasks.Load())
}

// RemoveTask removes the task with the specified ID from this pool.
func (tp *Pool) RemoveTask(taskID string) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()

	delete(tp.tasks, taskID)
}

// HandleTaskExit processes task completion and determines if the worker should continue.
func (tp *Pool) HandleTaskExit(isReplacementThread bool) (canContinue bool) {
	// Global queues do not finish and cannot be waited on. If this is NOT a global queue,
	// we check if we are empty and finished, and if so, wake any waiters.
	if !tp.treatAsGlobal {
		uncompletedTasks := tp.totalTasks.Load() - tp.completedTasks.Load()

		if uncompletedTasks == 0 {
			tp.handlePoolDeconstruction()
		}
	}

	// If this is a replacement task, and we have more workers than the target for the pool, we exit
	if isReplacementThread && tp.workerPool.currentWorkers.Load() > tp.workerPool.maxWorkers.Load() {
		// Important to decrement alive workers inside the exitLock, so
		// we don't have multiple workers exiting when we only need the 1
		return false
	}

	// If we have already begun running the task,
	// we must finish and clean up before checking exitFlag again here.
	// The task *could* be canceled to speed things up, but that
	// is not our job.
	select {
	case <-tp.workerPool.ctx.Done():
		return false
	default:
		return true
	}
}

// GetRootPool returns the root task pool in the hierarchy.
func (tp *Pool) GetRootPool() *Pool {
	if tp.IsRoot() {
		return tp
	}

	tmpTp := tp
	for !tmpTp.parentTaskPool.IsRoot() {
		tmpTp = tmpTp.parentTaskPool
	}

	return tmpTp
}

// Status returns the current status of the task pool including completion progress.
func (tp *Pool) Status() PoolStatus {
	complete := tp.completedTasks.Load()
	total := tp.totalTasks.Load()

	progress := (float64(complete * 100)) / float64(total)
	if math.IsNaN(progress) {
		progress = 0
	}

	tp.taskLock.RLock()
	errorCount := len(tp.erroredTasks)
	tp.taskLock.RUnlock()

	return PoolStatus{
		Complete: complete,
		Failed:   errorCount,
		Total:    total,
		Progress: progress,
		Runtime:  time.Since(tp.createdAt),
	}
}

// Wait Parks the thread on the work queue until all the tasks have been queued and finish.
// **If you never call tp.SignalAllQueued(), the waiters will never wake up**
// Make sure that you SignalAllQueued before parking here if it is the only thread
// loading tasks.
// If you are parking a thread that is currently executing a task, you can
// pass that task in as well, and that task will also listen for exit events.
func (tp *Pool) Wait(supplementWorker bool, task ...*Task) {
	// Waiting on global queues does not make sense, they are not meant to end
	// or
	// All the tasks were queued, and they have all finished,
	// so no need to wait, we can "wake up" instantly.
	if tp.treatAsGlobal || (tp.allQueuedFlag.Load() && tp.totalTasks.Load()-tp.completedTasks.Load() == 0) {
		return
	}

	// If we want to park another thread that is currently executing a task,
	// e.g a directory scan waiting for the child file scans to complete,
	// we want to add a worker to the pool temporarily to supplement this one
	if supplementWorker {
		ctx, cancel := context.WithCancel(tp.workerPool.ctx)
		defer cancel()

		tp.workerPool.busyCount.Add(-1)
		tp.workerPool.addReplacementWorker(ctx)

		defer func() {
			tp.workerPool.busyCount.Add(1)
			tp.workerPool.removeWorker()
		}()
	}

	tp.log.Trace().Func(func(e *zerolog.Event) {
		_, file, line, _ := runtime.Caller(2)
		e.Msgf("Parking on worker pool from %s:%d", file, line)
	})

	if !tp.allQueuedFlag.Load() {
		tp.log.Warn().Msg("Going to sleep on pool without allQueuedFlag set! This task pool may never wake up!")
	}

	for _, t := range task {
		t.SetQueueState(Sleeping)
		t.SetResult(Result{
			"waiting": true,
		})
	}

	defer func() {
		for _, t := range task {
			t.SetQueueState(Executing)
			t.SetResult(Result{
				"waiting": false,
			})
		}
	}()

	tp.waiterCount.Add(1)
	defer tp.waiterCount.Add(-1)

	if len(task) != 0 {
		select {
		case <-tp.createdBy.Ctx.Done():
		case <-task[0].Ctx.Done():
		case <-tp.waiterGate:
		}
	} else {
		select {
		case <-tp.createdBy.Ctx.Done():
		case <-tp.waiterGate:
		}
	}

	tp.log.Trace().Func(func(e *zerolog.Event) {
		_, file, line, _ := runtime.Caller(2)
		e.Msgf("Woke up, returning to %s:%d", file, line)
	})
}

// LockExit acquires the exit lock for this task pool.
func (tp *Pool) LockExit() {
	tp.exitLock.Lock()
}

// UnlockExit releases the exit lock for this task pool.
func (tp *Pool) UnlockExit() {
	tp.exitLock.Unlock()
}

// AddError adds a task that encountered an error to the pool's error list.
func (tp *Pool) AddError(t *Task) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()

	tp.erroredTasks = append(tp.erroredTasks, t)
}

// AddCleanup registers a cleanup function to run when the pool completes.
func (tp *Pool) AddCleanup(fn PoolCleanupFunc) {
	tp.exitLock.Lock()
	defer tp.exitLock.Unlock()

	if tp.allQueuedFlag.Load() && tp.completedTasks.Load() == tp.totalTasks.Load() && tp.cleanupsDone.Load() {
		// Caller expects `AddCleanup` to execute asynchronously, so we must run the
		// cleanup function in its own go routine
		go fn(tp)

		return
	}

	tp.cleanupFn = append(tp.cleanupFn, fn)
}

// Errors returns the list of tasks that encountered errors in this pool.
func (tp *Pool) Errors() []*Task {
	return tp.erroredTasks
}

// Cancel cancels all tasks in this pool.
func (tp *Pool) Cancel() {
	// Dont allow more tasks to join the queue while we are canceling them
	tp.taskLock.Lock()

	tp.allQueuedFlag.Store(true)

	for _, t := range tp.tasks {
		t.Cancel()
	}

	tp.taskLock.Unlock()
}

// QueueTask adds a task to this pool for execution.
func (tp *Pool) QueueTask(tsk *Task) (err error) {
	select {
	case <-tp.workerPool.ctx.Done():
		tp.log.Warn().Msg("Not queuing task while worker pool is going down")

		return err
	default:
	}

	if tsk.err != nil {
		// Tasks that have failed will not be re-tried. If the errored task is removed from the
		// task map, then it will be re-tried because the previous error was lost. This can be
		// sometimes be useful, some tasks auto-remove themselves after they finish.
		tp.log.Warn().Msg("Not re-queueing task that has error set, please restart weblens to try again")

		return err
	}

	if tsk.taskPool != nil && (tsk.taskPool != tp || tsk.queueState != Created) {
		// Task is already queued, we are not allowed to move it to another queue.
		// We can call .ClearAndRecompute() on the task and it will queue it
		// again, but it cannot be transferred
		if tsk.taskPool != tp {
			tp.log.Warn().Msgf("Attempted to re-queue a [%s] task that is already in a queue", tsk.jobName)

			return err
		}

		tsk.taskPool.tasks[tsk.taskID] = tsk

		return err
	}

	if tp.allQueuedFlag.Load() {
		// We cannot add tasks to a queue that has been closed
		return wlerrors.WithStack(wlerrors.New("attempting to add task to closed task queue"))
	}

	tp.totalTasks.Add(1)

	if tp.parentTaskPool != nil {
		tmpTp := tp
		for tmpTp.parentTaskPool != nil {
			tmpTp = tmpTp.parentTaskPool
		}

		if tmpTp != tp {
			tmpTp.totalTasks.Add(1)
		}
	}

	// Set the tasks queue
	tsk.taskPool = tp

	tp.workerPool.lifetimeQueuedCount.Add(1)

	// Put the task in the queue
	tsk.queueState = InQueue
	if len(tp.workerPool.retryBuffer) != 0 || len(tp.workerPool.taskStream) == cap(tp.workerPool.taskStream) {
		tp.workerPool.addToRetryBuffer(tsk)
	} else {
		tp.workerPool.taskStream <- tsk
	}

	tsk.taskPool.tasks[tsk.taskID] = tsk

	return err
}

// MarkGlobal specifies the work queue as being a "global" one
func (tp *Pool) MarkGlobal() {
	tp.treatAsGlobal = true
}

// IsGlobal returns true if this task pool is treated as a global pool.
func (tp *Pool) IsGlobal() bool {
	return tp.treatAsGlobal
}

// CreatedInTask returns the task that created this pool, or nil if created outside a task.
func (tp *Pool) CreatedInTask() *Task {
	if tp.createdBy == nil {
		return nil
	}

	return tp.createdBy
}

// SignalAllQueued marks that all tasks have been queued to this pool.
func (tp *Pool) SignalAllQueued() {
	if tp.treatAsGlobal {
		tp.log.Error().Msg("Attempt to signal all queued for global queue")
	}

	if tp.allQueuedFlag.Load() {
		tp.log.Warn().Msg("Trying to signal all queued on already all-queued task pool")

		return
	}

	tp.exitLock.Lock()
	// If all tasks finish (e.g. early failure, etc.) before we signal that they are all queued,
	// the final exiting task will not let the waiters out, so we must do it here. We must also
	// remove the task pool from the worker pool for the same reason
	if tp.completedTasks.Load() == tp.totalTasks.Load() {
		tp.log.Debug().Msgf("SignalAllQueued: Task pool [%s] already complete, waking up waiters and removing task pool from worker pool", tp.ID())

		tp.handlePoolDeconstruction()
	} else {
		tp.log.Trace().Msg("Task Pool NOT Complete")
	}

	tp.allQueuedFlag.Store(true)
	tp.exitLock.Unlock()

	if tp.hasQueueThread {
		tp.workerPool.removeWorker()
		tp.hasQueueThread = false
	}
}

func (tp *Pool) handlePoolDeconstruction() {
	// Check if the task pool is still active in the worker pool. If it is not,
	// we do not attempt to wake the waiters or remove it again
	poolStillActive := tp.workerPool.GetTaskPool(tp.ID()) != nil

	// Even if we are out of tasks, if we have not been told all tasks
	// were queued, we do not wake the waiters
	if poolStillActive && tp.allQueuedFlag.Load() {
		tp.log.Debug().Msgf("All tasks completed, closing task pool [%s]", tp.ID())

		close(tp.waiterGate)

		// Check if all waiters have awoken before closing the queue, spin and sleep for 1ms if not
		// Should only loop a handful of times, if at all, we just need to wait for the waiters to
		// lock and then release immediately. Should take, like, microseconds (?) each
		for tp.waiterCount.Load() != 0 {
			time.Sleep(time.Millisecond * 1)
		}

		// Once all the tasks have exited, this worker pool is now closing, and so we must run
		// its cleanup routine(s)
		if len(tp.cleanupFn) != 0 {
			tp.runCleanups()
		}

		tp.workerPool.removeTaskPool(tp.ID())
	}
}

func (tp *Pool) addTask(task *Task) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()

	tp.tasks[task.taskID] = task
}

func (tp *Pool) runCleanups() {
	defer func() {
		if r := recover(); r != nil {
			tp.log.Error().Stack().Err(wlerrors.New(fmt.Sprint(r))).Msg("Failed to execute taskPool cleanup")
		}

		tp.cleanupsDone.Store(true)
	}()

	for _, fn := range tp.cleanupFn {
		fn(tp)
	}
}
