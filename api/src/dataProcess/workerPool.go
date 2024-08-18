package dataProcess

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/util"
	"github.com/ethrousseau/weblens/api/util/wlog"
	"github.com/google/uuid"

	"github.com/ethrousseau/weblens/api/types"
)

type hit struct {
	time   time.Time
	target *task
}

type workChannel chan *task

type hitChannel chan hit

type workerPool struct {
	maxWorkers     *atomic.Int64 // Max allowed worker count
	currentWorkers *atomic.Int64 // Current total of workers on the pool
	busyCount      *atomic.Int64 // Number of workers currently executing a task

	lifetimeQueuedCount *atomic.Int64

	taskMu sync.Mutex
	taskMap map[types.TaskId]types.Task

	poolMu sync.Mutex
	poolMap map[types.TaskId]types.TaskPool

	taskStream   workChannel
	taskBufferMu sync.Mutex
	retryBuffer  []*task
	hitStream    hitChannel

	exitFlag *atomic.Int64
}

func NewWorkerPool(initWorkers int) (types.WorkerPool, types.TaskPool) {
	if initWorkers == 0 {
		initWorkers = 1
	}
	wlog.Info.Printf("Starting new worker pool with %d workers", initWorkers)

	newWp := &workerPool{
		maxWorkers:          &atomic.Int64{},
		currentWorkers:      &atomic.Int64{},
		busyCount:           &atomic.Int64{},
		lifetimeQueuedCount: &atomic.Int64{},
		exitFlag:            &atomic.Int64{},

		taskMu: sync.Mutex{},
		taskMap: map[types.TaskId]types.Task{},

		poolMu: sync.Mutex{},
		poolMap: map[types.TaskId]types.TaskPool{},

		taskStream:   make(workChannel, initWorkers*1000),
		retryBuffer:  []*task{},
		taskBufferMu: sync.Mutex{},

		hitStream: make(hitChannel, initWorkers*2),
	}
	// Worker pool starts disabled
	newWp.exitFlag.Store(1)

	newWp.maxWorkers.Add(int64(initWorkers))

	// Worker pool always has one global queue
	globalPool := newWp.newTaskPoolInternal()
	globalPool.id = "GLOBAL"
	globalPool.MarkGlobal()

	return newWp, globalPool
}

// NewTaskPool `replace` spawns a temporary replacement thread on the parent worker pool.
// This prevents a deadlock when the queue fills up while adding many tasks, and none are being executed
//
// `parent` allows chaining of task pools for floating updates to the top. This makes
// it possible for clients to subscribe to a single task, and get notified about
// all of the sub-updates of that task
// See taskPool.go
func (wp *workerPool) NewTaskPool(replace bool, createdBy types.Task) types.TaskPool {
	tp := wp.newTaskPoolInternal()
	if createdBy != nil {
		tp.createdBy = createdBy
		if !createdBy.GetTaskPool().IsGlobal() {
			tp.createdBy = createdBy
		}
	}
	if replace {
		wp.addReplacementWorker()
		tp.hasQueueThread = true
	}

	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	wp.poolMap[tp.ID()] = tp

	return tp
}

func (wp *workerPool) GetTaskPool(tpId types.TaskId) types.TaskPool {
	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	return wp.poolMap[tpId]
}

func (wp *workerPool) removeTaskPool(tpId types.TaskId) {
	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	delete(wp.poolMap, tpId)
}

func (wp *workerPool) GetTaskPoolByTaskType(taskType types.TaskType) types.TaskPool {
	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	for _, tp := range wp.poolMap {
		if tp.CreatedInTask() != nil && tp.CreatedInTask().TaskType() == taskType {
			return tp
		}
	}
	return nil
}

func workerRecover(task *task, workerId int64) {
	err := recover()
	if err != nil {
		// Make sure what we got is an error
		switch err.(type) {
		case error:
			if err.(error).Error() == ErrTaskExit.Error() || err.(error).Error() == ErrTaskError.Error() {
				return
			}
		default:
			err = errors.New(fmt.Sprint(err))
		}

		wlog.ErrorCatcher.Printf("Worker %d recovered error: %s\n%s\n", workerId, err, debug.Stack())
		task.error(err.(error))
	}
}

// saftyWork wraps the task execution with a recover, so if there are any panics
// during the task, we can catch them, display them, and safely remove the task.
func safetyWork(task *task, workerId int64) {
	defer workerRecover(task, workerId)
	task.work(task)

	task.updateMu.Lock()
	if task.postAction != nil && task.exitStatus == TaskSuccess {
		task.updateMu.Unlock()
		task.postAction(task.result)
		return
	}
	task.updateMu.Unlock()
}

func (wp *workerPool) AddHit(time time.Time, target types.Task) {
	wp.hitStream <- hit{time: time, target: target.(*task)}
}

// The reaper handles timed cancellation of tasks. If a task might
// wait on some action from the client, it should queue a timeout with
// the reaper so it doesn't hang forever
func (wp *workerPool) reaper() {
	timerStream := make(chan *task)
	for wp.exitFlag.Load() == 0 {
		select {
		case newHit := <-wp.hitStream:
			go func(h hit) { time.Sleep(time.Until(h.time)); timerStream <- h.target }(newHit)
		case task := <-timerStream:
			// Possible that the task has kicked it's timeout down the road
			// since it first queued it, we must check that we are past the
			// timeout before cancelling the task. Also check that it
			// has not already finished
			if task.queueState != Exited && time.Until(task.timeout) <= 0 && task.timeout.Unix() != 0 {
				wlog.Warning.Printf("Sending timeout signal to T[%s]\n", task.taskId)
				task.Cancel()
				task.error(ErrTaskTimeout)
			}
		}
	}
}

func (wp *workerPool) bufferDrainer() {
	for wp.exitFlag.Load() == 0 {
		wp.taskBufferMu.Lock()
		if len(wp.retryBuffer) != 0 && len(wp.taskStream) == 0 {
			for _, t := range wp.retryBuffer {
				wp.taskStream <- t
			}
			wp.retryBuffer = []*task{}
		}
		wp.taskBufferMu.Unlock()
		time.Sleep(time.Second * 10)
	}

	wlog.ErrTrace(errors.New("buffer drainer exited"))
}

func (wp *workerPool) addToRetryBuffer(tasks ...*task) {
	wp.taskBufferMu.Lock()
	defer wp.taskBufferMu.Unlock()
	// util.Debug.Printf("Adding %d tasks to the buffer!", len(tasks))
	wp.retryBuffer = append(wp.retryBuffer, tasks...)
}

// Run launches the standard threads for this worker pool
func (wp *workerPool) Run() {
	wp.exitFlag.Store(0)

	// Spawn the timeout checker
	util.LabelThread(
		func(_ context.Context) {
			go wp.reaper()
		}, "", "Reaper",
	)

	// Spawn the buffer worker
	util.LabelThread(
		func(_ context.Context) {
			go wp.bufferDrainer()
		}, "", "Buffer Drainer",
	)


	// Spawn the status printer
	util.LabelThread(
		func(_ context.Context) {
			go wp.statusReporter()
		}, "", "Status Reporter",
	)


	var i int64
	for i = 0; i < wp.maxWorkers.Load(); i++ {
		// These are the base, 'omnipresent' threads for this pool,
		// so they are NOT replacement workers. See wp.execWorker
		// for more info
		wp.execWorker(false)
	}
}

func (wp *workerPool) GetTask(taskId types.TaskId) types.Task {
	wp.taskMu.Lock()
	t := wp.taskMap[taskId]
	wp.taskMu.Unlock()
	return t
}

func (wp *workerPool) addTask(task types.Task) {
	wp.taskMu.Lock()
	defer wp.taskMu.Unlock()
	wp.taskMap[task.TaskId()] = task
}

func (wp *workerPool) removeTask(taskId types.TaskId) {
	wp.taskMu.Lock()
	t := wp.taskMap[taskId]
	wp.taskMu.Unlock()
	if t == nil {
		return
	}

	t.GetTaskPool().RemoveTask(taskId)

	wp.taskMu.Lock()
	defer wp.taskMu.Unlock()
	delete(wp.taskMap, taskId)
}

// For debugging only. This does nothing if the DEV_MODE env var is set to false, as no stopwatch results
// will be printed reguardless
const printStopwatchResults = false

// Main worker method, spawn a worker and loop over the task channel
//
// `replacement` specifies if the worker is a temporary replacement for another
// task that is parking for a long time. Replacement workers behave a bit
// different to minimize parked time of the other task.
func (wp *workerPool) execWorker(replacement bool) {
	wp.currentWorkers.Add(1)
	util.LabelThread(
		func(_ context.Context) {
			go func(workerId int64) {

				// Inc alive workers
				defer wp.currentWorkers.Add(-1)
				// util.Debug.Printf("Worker %d reporting for duty o7", workerId)

			WorkLoop:
				for t := range wp.taskStream {

					// Even if we get a new task, if the wp is marked for exit, we just exit
					if wp.exitFlag.Load() == 1 {
						// We don't care about the exitLock here, since the whole wp
						// is going down anyway.

						break WorkLoop
					}

					// Replacement workers are not allowed to do "scan_directory" tasks
					if replacement && t.taskType == ScanDirectoryTask {

						// If there are twice the number of free spaces in the chan, don't bother pulling
						// everything into the waiting buffer, just put it at the end right now.
						if cap(wp.taskStream)-len(wp.taskStream) > int(wp.currentWorkers.Load())*2 {
							// util.Debug.Println("Replacement worker putting scan dir task back")
							wp.taskStream <- t
							continue
						}
						tBuf := []*task{t}
						for t = range wp.taskStream {
							if t.taskType == ScanDirectoryTask {
								tBuf = append(tBuf, t)
							} else {
								break
							}
						}
						wp.addToRetryBuffer(tBuf...)
					}

					// Inc tasks being processed
					wp.busyCount.Add(1)
					t.SwLap("Task start")
					wlog.Debug.Printf("Starting %s task T[%s]", t.taskType, t.taskId)
					start := time.Now()
					util.LabelThread(
						func(_ context.Context) {
							safetyWork(t, workerId)
						}, "Worker Id", strconv.FormatInt(workerId, 10), "Task", string(t.taskType),
					)
					wlog.Debug.Printf("Finished %s task T[%s] in %s", t.taskType, t.taskId, time.Since(start))
					t.SwLap("Task finish")
					// Dec tasks being processed
					wp.busyCount.Add(-1)

					// Wake any waiters on this task
					t.waitMu.Unlock()

					// Tasks must set their completed flag before exiting
					// if it wasn't done in the work body, we do it for them
					if t.queueState != Exited {
						t.success("closed by worker pool")
					}

					// Notify this task has completed, and unsubscribe any subscribers to it
					result := t.GetResults()
					result["task_id"] = t.taskId
					result["exit_status"] = t.exitStatus
					wp.poolMu.Lock()
					var complete int64
					for _, p := range wp.poolMap {
						status := p.Status()
						complete += status.Complete
					}
					wp.poolMu.Unlock()
					result["queue_remaining"] = complete
					result["queue_total"] = wp.lifetimeQueuedCount.Load()

					t.caster.PushTaskUpdate(
						t, TaskCompleteEvent,
						result,
					)
					t.caster.UnsubTask(t)

					/* No websocket updates to the client past this point */

					// A common use for the cleanup method is to close a caster,
					// so we must do this after we push the final update
					if t.cleanup != nil {
						t.cleanup()
						t.cleanup = nil
					}

					// Potentially find the task pool that houses this task pool. All child
					// task pools report their status to the root task pool as well.
					// Do not use any global pool as the root
					rootTaskPool := t.taskPool.GetRootPool().(*taskPool)

					if printStopwatchResults {
						t.sw.PrintResults(true)
					}
					if t.exitStatus == TaskError && !rootTaskPool.IsGlobal() {
						rootTaskPool.AddError(t)
						// util.Warning.Printf("T[%s] exited with error status: %s\n", t.taskId, t.exitStatus)
					}

					if !t.persistent {
						wp.removeTask(t.taskId)
					}

					canContinue := true
					directParent := t.GetTaskPool().(*taskPool)
					if directParent.IsRoot() {
						// Updating the number of workers and then checking it's value is dangerous
						// to do concurrently. Specifically, the waiterGate lock on the queue will,
						// very rarely, attempt to unlock twice if two tasks finish at the same time.
						// So we must treat this whole area as a critical section
						directParent.LockExit()

						// Set values and notifications now that task has completed. Returns
						// a bool that specifies if this thread should continue and grab another
						// task, or if it should exit
						canContinue = directParent.handleTaskExit(replacement)

						directParent.UnlockExit()
					} else {
						// Must hold both locks (and must acquire root lock first) to enter a dual-update.
						// Any other ordering will result in race conditions or deadlocks
						rootTaskPool.LockExit()

						directParent.LockExit()
						uncompletedTasks := directParent.totalTasks.Load() - directParent.completedTasks.Load()
						wlog.Debug.Printf(
							"Uncompleted tasks on tp created by %s: %d",
							directParent.CreatedInTask().TaskId(), uncompletedTasks-1,
						)
						canContinue = directParent.handleTaskExit(replacement)

						// We *should* get the same canContinue value from here, so we do not
						// check it a second time. If we *don't* get the same value, we can safely ignore it
						rootTaskPool.handleTaskExit(replacement)

						directParent.UnlockExit()
						rootTaskPool.UnlockExit()
					}

					if !canContinue {
						break WorkLoop
					}

				}
			}(wp.currentWorkers.Load())
		}, "Worker Id", strconv.FormatInt(wp.currentWorkers.Load(), 10), "Task", "Idle",
	)
}

/*
We can create "replacement" workers as to not create a deadlock.
e.g. all n `wp.maxWorkers` are "scan_directory" tasks parked and waiting
for their "scan_file" children tasks to complete, but never will since
all worker threads are taken up by blocked tasks.

Replacement workers:
1. Are not allowed to pick up "scan_directory" tasks
2. will exit once the main thread has woken up, and shrunk the worker pool back down
*/
func (wp *workerPool) addReplacementWorker() {
	wp.maxWorkers.Add(1)
	wp.execWorker(true)
}

func (wp *workerPool) removeWorker() {
	wp.maxWorkers.Add(-1)
}

func (wp *workerPool) newTaskPoolInternal() *taskPool {
	tpId, err := uuid.NewUUID()
	if err != nil {
		wlog.ShowErr(err)
		return nil
	}

	newQueue := &taskPool{
		id:             types.TaskId(tpId.String()),
		totalTasks:     &atomic.Int64{},
		completedTasks: &atomic.Int64{},
		waiterCount:    &atomic.Int32{},
		waiterGate: sync.Mutex{},
		exitLock:   sync.Mutex{},
		tasks:          map[types.TaskId]*task{},
		taskLock:   sync.Mutex{},
		workerPool:     wp,
		erroredTasks:   make(chan *task, 1000),
		createdAt:      time.Now(),
	}

	// The waiterGate begins locked, and will only unlock when all
	// tasks have been queued and finish, then the waiters are released
	newQueue.waiterGate.Lock()

	return newQueue
}

// Status returns the count of tasks in the queue, the total number of tasks accepted,
// number of busy workers, and the total number of live workers in the worker pool
func (wp *workerPool) Status() (int, int, int, int, int) {
	total := int(wp.lifetimeQueuedCount.Load())
	wp.taskBufferMu.Lock()
	defer wp.taskBufferMu.Unlock()

	return len(wp.taskStream), total, int(wp.busyCount.Load()), int(wp.currentWorkers.Load()), len(wp.retryBuffer)
}

func (wp *workerPool) statusReporter() {
	var lastCount int
	var waitTime time.Duration = 1
	for {
		time.Sleep(time.Second * waitTime)
		remaining, total, busy, alive, retrySize := wp.Status()
		if lastCount != remaining {
			lastCount = remaining
			waitTime = 1
			wlog.Info.Printf(
				"Task worker pool status : Queued[%d]/Total[%d], Buffered[%d], Busy[%d], Alive[%d]",
				remaining, total, retrySize, busy, alive,
			)
		} else if waitTime < time.Second*10 {
			waitTime += 1
		}
	}
}
