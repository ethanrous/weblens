package task

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
)

type TaskPool struct {
	id Id

	treatAsGlobal  bool
	hasQueueThread bool

	tasks    map[Id]*Task
	taskLock sync.RWMutex

	totalTasks     atomic.Int64
	completedTasks atomic.Int64
	waiterCount    atomic.Int32
	waiterGate     chan struct{}
	exitLock       sync.Mutex

	workerPool     *WorkerPool
	parentTaskPool *TaskPool
	createdBy      *Task
	createdAt      time.Time

	cleanupFn func(pool Pool)

	allQueuedFlag atomic.Bool

	erroredTasks []*Task
}

func (tp *TaskPool) IsRoot() bool {
	if tp == nil {
		return false
	}
	return tp.parentTaskPool == nil || tp.parentTaskPool.IsGlobal()
}

func (tp *TaskPool) GetWorkerPool() *WorkerPool {
	return tp.workerPool
}

func (tp *TaskPool) ID() Id {
	return tp.id
}

func (tp *TaskPool) addTask(task *Task) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()
	tp.tasks[task.taskId] = task
}

func (tp *TaskPool) RemoveTask(taskId Id) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()
	delete(tp.tasks, taskId)

}

func (tp *TaskPool) handleTaskExit(replacementThread bool) (canContinue bool) {

	// Global queues do not finish and cannot be waited on. If this is NOT a global queue,
	// we check if we are empty and finished, and if so, wake any waiters.
	if !tp.treatAsGlobal {
		uncompletedTasks := tp.totalTasks.Load() - tp.completedTasks.Load()

		// Even if we are out of tasks, if we have not been told all tasks
		// were queued, we do not wake the waiters
		if uncompletedTasks == 0 && tp.allQueuedFlag.Load() {
			if tp.waiterCount.Load() != 0 {
				log.Trace.Println("Pool complete, waking sleepers!")
				// TODO - move pool completion to cleanup function
				// if tp.createdBy != nil && tp.createdBy.caster != nil {
				// 	tp.createdBy.caster.PushPoolUpdate(tp, websocket.PoolCompleteEvent, nil)
				// }
				close(tp.waiterGate)

				// Check if all waiters have awoken before closing the queue, spin and sleep for 10ms if not
				// Should only loop a handful of times, if at all, we just need to wait for the waiters to
				// lock and then release immediately. Should take, like, nanoseconds (?) each
				for tp.waiterCount.Load() != 0 {
					time.Sleep(time.Millisecond * 1)
				}
			}

			// Once all the tasks have exited, this worker pool is now closing, and so we must run
			// its cleanup routine(s)
			if tp.cleanupFn != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.ShowErr(errors.New(fmt.Sprint(r)), "Failed to execute taskPool cleanup")
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
		tp.workerPool.currentWorkers.Add(-1)

		return false
	}

	// If we have already began running the task,
	// we must finish and clean up before checking exitFlag again here.
	// The task *could* be cancelled to speed things up, but that
	// is not our job.
	if tp.workerPool.exitFlag.Load() == 1 {
		// Dec alive workers
		tp.workerPool.currentWorkers.Add(-1)
		return false
	}

	return true
}

func (tp *TaskPool) GetRootPool() *TaskPool {
	if tp.IsRoot() {
		return tp
	}

	tmpTp := tp
	for !tmpTp.parentTaskPool.IsRoot() {
		tmpTp = tmpTp.parentTaskPool
	}
	return tmpTp
}

func (tp *TaskPool) Status() PoolStatus {
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
func (tp *TaskPool) Wait(supplementWorker bool, task ...*Task) {
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

	_, file, line, _ := runtime.Caller(1)
	log.Trace.Func(func(l log.Logger) { l.Printf("Parking on worker pool from %s:%d\n", file, line) })

	if !tp.allQueuedFlag.Load() {
		log.Warning.Println("Going to sleep on pool without allQueuedFlag set! This thread may never wake up!")
	}

	tp.waiterCount.Add(1)
	if task[0] != nil {
		select {
		case <-task[0].signalChan:
		case <-tp.waiterGate:
		}
	} else {
		<-tp.waiterGate
	}
	tp.waiterCount.Add(-1)

	log.Trace.Func(func(l log.Logger) { l.Printf("Woke up, returning to %s:%d\n", file, line) })

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

func (tp *TaskPool) AddError(t *Task) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()
	tp.erroredTasks = append(tp.erroredTasks, t)
}

func (tp *TaskPool) AddCleanup(fn func(pool Pool)) {
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

func (tp *TaskPool) Errors() []*Task {
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

func (tp *TaskPool) QueueTask(t *Task) (err error) {

	if tp.workerPool.exitFlag.Load() == 1 {
		log.Warning.Println("Not queuing task while worker pool is going down")
		return
	}

	if t.err != nil {
		// Tasks that have failed will not be re-tried. If the errored task is removed from the
		// task map, then it will be re-tried because the previous error was lost. This can be
		// sometimes be useful, some tasks auto-remove themselves after they finish.
		log.Warning.Println("Not re-queueing task that has error set, please restart weblens to try again")
		return
	}

	if t.taskPool != nil && (t.taskPool != tp || t.queueState != PreQueued) {
		// Task is already queued, we are not allowed to move it to another queue.
		// We can call .ClearAndRecompute() on the task and it will queue it
		// again, but it cannot be transferred
		if t.taskPool != tp {
			log.Warning.Printf("Attempted to re-queue a [%s] task that is already in a queue", t.jobName)
			return
		}
		t.taskPool.tasks[t.taskId] = t
		return
	}

	if tp.allQueuedFlag.Load() {
		// We cannot add tasks to a queue that has been closed
		return werror.WithStack(errors.New("attempting to add task to closed task queue"))
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

func (tp *TaskPool) CreatedInTask() *Task {
	return tp.createdBy
}

func (tp *TaskPool) SignalAllQueued() {
	if tp.treatAsGlobal {
		log.Error.Println("Attempt to signal all queued for global queue")
	}
	if tp.allQueuedFlag.Load() {
		log.Warning.Println("Trying to signal all queued on already all-queued task pool")
		return
	}

	tp.exitLock.Lock()
	// If all tasks finish (e.g. early failure, etc.) before we signal that they are all queued,
	// the final exiting task will not let the waiters out, so we must do it here. We must also
	// remove the task pool from the worker pool for the same reason
	if tp.completedTasks.Load() == tp.totalTasks.Load() {
		close(tp.waiterGate)
		tp.workerPool.removeTaskPool(tp.ID())
	}
	tp.allQueuedFlag.Store(true)
	tp.exitLock.Unlock()

	if tp.hasQueueThread {
		tp.workerPool.removeWorker()
		tp.hasQueueThread = false
	}
}

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

type Pool interface {
	ID() Id

	QueueTask(*Task) error

	MarkGlobal()
	IsGlobal() bool
	SignalAllQueued()

	CreatedInTask() *Task

	Wait(bool, ...*Task)
	Cancel()

	IsRoot() bool
	Status() PoolStatus
	AddCleanup(fn func(Pool))

	GetRootPool() *TaskPool
	GetWorkerPool() *WorkerPool

	LockExit()
	UnlockExit()

	RemoveTask(Id)

	// NotifyTaskComplete(Task, websocket.BroadcasterAgent, ...any)

	// ScanDirectory(WeblensFile, websocket.BroadcasterAgent) Task
	// ScanFile(WeblensFile, websocket.BroadcasterAgent) Task
	// WriteToFile(FileId, int64, int64, websocket.BroadcasterAgent) Task
	// MoveFile(FileId, FileId, string, FileEvent, websocket.BroadcasterAgent) Task
	// GatherFsStats(WeblensFile, websocket.BroadcasterAgent) Task
	// Backup(InstanceId, websocket.BroadcasterAgent) Task
	// HashFile(WeblensFile, websocket.BroadcasterAgent) Task
	// CreateZip(files []WeblensFile, username Username, shareId ShareId, casters websocket.BroadcasterAgent) Task
	// CopyFileFromCore(WeblensFile, Client, websocket.BroadcasterAgent) Task

	Errors() []*Task
	AddError(t *Task)
}
