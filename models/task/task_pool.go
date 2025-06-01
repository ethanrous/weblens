package task

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/modules/errors"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/rs/zerolog"
)

var _ task_mod.Pool = (*TaskPool)(nil)

type TaskPool struct {
	allQueuedFlag  atomic.Bool
	cleanupFn      task_mod.PoolCleanupFunc
	completedTasks atomic.Int64
	createdAt      time.Time
	createdBy      *Task
	erroredTasks   []task_mod.Task
	exitLock       sync.Mutex
	hasQueueThread bool
	id             string
	log            zerolog.Logger
	parentTaskPool *TaskPool
	taskLock       sync.RWMutex
	tasks          map[string]*Task
	totalTasks     atomic.Int64
	treatAsGlobal  bool
	waiterCount    atomic.Int32
	waiterGate     chan struct{}
	workerPool     *WorkerPool
}

func (tp *TaskPool) IsRoot() bool {
	if tp == nil {
		return false
	}
	return tp.parentTaskPool == nil || tp.parentTaskPool.IsGlobal()
}

func (tp *TaskPool) GetWorkerPool() task_mod.WorkerPool {
	return tp.workerPool
}

func (tp *TaskPool) ID() string {
	return tp.id
}

func (tp *TaskPool) IncTaskCount(count int) {
	tp.totalTasks.Add(int64(count))
}

func (tp *TaskPool) GetTotalTaskCount() int {
	return int(tp.totalTasks.Load())
}

func (tp *TaskPool) IncCompletedTasks(count int) {
	tp.completedTasks.Add(int64(count))
}

func (tp *TaskPool) GetCompletedTaskCount() int {
	return int(tp.completedTasks.Load())
}

func (tp *TaskPool) addTask(task *Task) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()
	tp.tasks[task.taskId] = task
}

func (tp *TaskPool) RemoveTask(taskId string) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()
	delete(tp.tasks, taskId)

}

func (tp *TaskPool) HandleTaskExit(replacementThread bool) (canContinue bool) {

	// Global queues do not finish and cannot be waited on. If this is NOT a global queue,
	// we check if we are empty and finished, and if so, wake any waiters.
	if !tp.treatAsGlobal {
		uncompletedTasks := tp.totalTasks.Load() - tp.completedTasks.Load()

		// Even if we are out of tasks, if we have not been told all tasks
		// were queued, we do not wake the waiters
		if uncompletedTasks == 0 && tp.allQueuedFlag.Load() {
			tp.log.Debug().Msg("All tasks completed, closing task pool")
			close(tp.waiterGate)

			// Check if all waiters have awoken before closing the queue, spin and sleep for 10ms if not
			// Should only loop a handful of times, if at all, we just need to wait for the waiters to
			// lock and then release immediately. Should take, like, nanoseconds (?) each
			for tp.waiterCount.Load() != 0 {
				time.Sleep(time.Millisecond * 1)
			}

			// Once all the tasks have exited, this worker pool is now closing, and so we must run
			// its cleanup routine(s)
			if tp.cleanupFn != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							tp.log.Error().Stack().Err(errors.New(fmt.Sprint(r))).Msg("Failed to execute taskPool cleanup")
						}
					}()
					tp.cleanupFn(tp)
				}()
			}
			tp.workerPool.removeTaskPool(tp.ID())
		}
	}
	// If this is a replacement task, and we have more workers than the target for the pool, we exit
	if replacementThread && tp.workerPool.currentWorkers.Load() > tp.workerPool.maxWorkers.Load() {
		// Important to decrement alive workers inside the exitLock, so
		// we don't have multiple workers exiting when we only need the 1
		// tp.workerPool.currentWorkers.Add(-1)

		return false
	}

	// If we have already begun running the task,
	// we must finish and clean up before checking exitFlag again here.
	// The task *could* be cancelled to speed things up, but that
	// is not our job.
	select {
	case <-tp.workerPool.ctx.Done():
		return false
	default:
		return true
	}
}

func (tp *TaskPool) GetRootPool() task_mod.Pool {
	if tp.IsRoot() {
		return tp
	}

	tmpTp := tp
	for !tmpTp.parentTaskPool.IsRoot() {
		tmpTp = tmpTp.parentTaskPool
	}
	return tmpTp
}

func (tp *TaskPool) Status() task_mod.PoolStatus {
	complete := tp.completedTasks.Load()
	total := tp.totalTasks.Load()
	progress := (float64(complete * 100)) / float64(total)
	if math.IsNaN(progress) {
		progress = 0
	}

	tp.taskLock.RLock()
	errorCount := len(tp.erroredTasks)
	tp.taskLock.RUnlock()

	return task_mod.PoolStatus{
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
func (tp *TaskPool) Wait(supplementWorker bool, task ...task_mod.Task) {
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
		tp.workerPool.busyCount.Add(-1)
		tp.workerPool.addReplacementWorker()
	}

	tp.log.Trace().Func(func(e *zerolog.Event) {
		_, file, line, _ := runtime.Caller(2)
		e.Msgf("Parking on worker pool from %s:%d", file, line)
	})

	if !tp.allQueuedFlag.Load() {
		tp.log.Warn().Msg("Going to sleep on pool without allQueuedFlag set! This task pool may never wake up!")
	}

	tp.waiterCount.Add(1)
	if len(task) != 0 {
		select {
		case <-tp.createdBy.Ctx.Done():
		case <-task[0].(*Task).Ctx.Done():
		case <-tp.waiterGate:
		}
	} else {
		select {
		case <-tp.createdBy.Ctx.Done():
		case <-tp.waiterGate:
		}
	}
	tp.waiterCount.Add(-1)

	tp.log.Trace().Func(func(e *zerolog.Event) {
		_, file, line, _ := runtime.Caller(2)
		e.Msgf("Woke up, returning to %s:%d", file, line)
	})

	if supplementWorker {
		tp.workerPool.busyCount.Add(1)
		tp.workerPool.removeWorker()
	}
}

func (tp *TaskPool) LockExit() {
	tp.exitLock.Lock()
}

func (tp *TaskPool) UnlockExit() {
	tp.exitLock.Unlock()
}

func (tp *TaskPool) AddError(t task_mod.Task) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()
	tp.erroredTasks = append(tp.erroredTasks, t)
}

func (tp *TaskPool) AddCleanup(fn task_mod.PoolCleanupFunc) {
	tp.exitLock.Lock()
	defer tp.exitLock.Unlock()
	if tp.allQueuedFlag.Load() && tp.completedTasks.Load() == tp.totalTasks.Load() {
		// Caller expects `AddCleanup` to execute asynchronously, so we must run the
		// cleanup function in its own go routine
		go fn(tp)
		return
	}

	tp.cleanupFn = fn
}

func (tp *TaskPool) Errors() []task_mod.Task {
	return tp.erroredTasks
}

func (tp *TaskPool) Cancel() {
	// Dont allow more tasks to join the queue while we are cancelling them
	tp.taskLock.Lock()

	tp.allQueuedFlag.Store(true)

	for _, t := range tp.tasks {
		t.Cancel()
	}
	tp.taskLock.Unlock()

	// TODO - move this to the cleanup function as well
	// Signal to the client that this pool has been canceled, so we can reflect
	// that in the UI
	// Caster.PushPoolUpdate(tp, websocket.PoolCancelledEvent, nil)

}

func (tp *TaskPool) QueueTask(task task_mod.Task) (err error) {
	t := task.(*Task)

	select {
	case <-tp.workerPool.ctx.Done():
		tp.log.Warn().Msg("Not queuing task while worker pool is going down")
		return
	default:
	}

	if t.err != nil {
		// Tasks that have failed will not be re-tried. If the errored task is removed from the
		// task map, then it will be re-tried because the previous error was lost. This can be
		// sometimes be useful, some tasks auto-remove themselves after they finish.
		tp.log.Warn().Msg("Not re-queueing task that has error set, please restart weblens to try again")
		return
	}

	if t.taskPool != nil && (t.taskPool != tp || t.queueState != Created) {
		// Task is already queued, we are not allowed to move it to another queue.
		// We can call .ClearAndRecompute() on the task and it will queue it
		// again, but it cannot be transferred
		if t.taskPool != tp {
			tp.log.Warn().Msgf("Attempted to re-queue a [%s] task that is already in a queue", t.jobName)
			return
		}
		t.taskPool.tasks[t.taskId] = t
		return
	}

	if tp.allQueuedFlag.Load() {
		// We cannot add tasks to a queue that has been closed
		return errors.WithStack(errors.New("attempting to add task to closed task queue"))
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
	t.taskPool = tp

	tp.workerPool.lifetimeQueuedCount.Add(1)

	// Put the task in the queue
	t.queueState = InQueue
	if len(tp.workerPool.retryBuffer) != 0 || len(tp.workerPool.taskStream) == cap(tp.workerPool.taskStream) {
		tp.workerPool.addToRetryBuffer(t)
	} else {
		tp.workerPool.taskStream <- t
	}

	t.taskPool.tasks[t.taskId] = t
	return
}

// MarkGlobal specifies the work queue as being a "global" one
func (tp *TaskPool) MarkGlobal() {
	tp.treatAsGlobal = true
}

func (tp *TaskPool) IsGlobal() bool {
	return tp.treatAsGlobal
}

func (tp *TaskPool) CreatedInTask() task_mod.Task {
	if tp.createdBy == nil {
		return nil
	}

	return tp.createdBy
}

func (tp *TaskPool) SignalAllQueued() {
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
		close(tp.waiterGate)
		tp.workerPool.removeTaskPool(tp.ID())
		tp.log.Debug().Msg("Task pool already complete")
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
