package dataProcess

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type hit struct {
	time   time.Time
	target *task
}

type workChannel chan *task

type hitChannel chan hit

type workerPool struct {
	maxWorkers     *atomic.Int64 // Max allowed worker count
	currentWorkers *atomic.Int64 // Current worker count
	busyCount      *atomic.Int64 // Number of workers currently executing a task

	lifetimeQueuedCount *atomic.Int64

	taskMu  *sync.Mutex
	taskMap map[types.TaskId]types.Task

	taskStream   workChannel
	taskBufferMu *sync.Mutex
	taskBuffer   []*task
	hitStream    hitChannel

	exitFlag int
}

func NewWorkerPool(initWorkers int) (types.WorkerPool, types.TaskPool) {
	if initWorkers == 0 {
		initWorkers = 1
	}
	util.Info.Printf("Starting new worker pool with %d workers", initWorkers)

	newWp := &workerPool{
		maxWorkers:          &atomic.Int64{},
		currentWorkers:      &atomic.Int64{},
		busyCount:           &atomic.Int64{},
		lifetimeQueuedCount: &atomic.Int64{},

		taskMu:  &sync.Mutex{},
		taskMap: map[types.TaskId]types.Task{},

		taskStream:   make(workChannel, initWorkers*1000),
		taskBuffer:   []*task{},
		taskBufferMu: &sync.Mutex{},

		hitStream: make(hitChannel, initWorkers*2),

		// Worker pool starts disabled
		exitFlag: 1,
	}

	newWp.maxWorkers.Add(int64(initWorkers))

	// Worker pool always has one global queue
	globalPool := newWp.NewVirtualTaskPool()
	globalPool.MarkGlobal()

	return newWp, globalPool
}

func workerRecover(task *task, workerId int64) {
	err := recover()
	if err != nil {
		// Make sure what we got is an error
		switch err.(type) {
		case error:
		default:
			err = errors.New(fmt.Sprint(err))
		}
		if err == ErrTaskExit || err == ErrTaskError {
			return
		}

		util.ErrorCatcher.Printf("Worker %d recovered error: %s\n%s\n", workerId, err, debug.Stack())
		task.error(err.(error))
	}
}

func safetyWork(task *task, workerId int64) {
	defer workerRecover(task, workerId)
	task.work(task)
}

func (wp *workerPool) AddHit(time time.Time, target types.Task) {
	wp.hitStream <- hit{time: time, target: target.(*task)}
}

// The reaper handles timed cancellation of tasks. If a task might
// wait on some action from the client, it should queue a timeout with
// the reaper so it doesn't hang forever
func (wp *workerPool) reaper() {
	timerStream := make(chan *task)
	for wp.exitFlag == 0 {
		select {
		case newHit := <-wp.hitStream:
			go func(h hit) { time.Sleep(time.Until(h.time)); timerStream <- h.target }(newHit)
		case task := <-timerStream:
			// Possible that the task has kicked it's timeout down the road
			// since it first queued it, we must check that we are past the
			// timeout before cancelling the task. Also check that it
			// has not already finished
			if task.queueState != Exited && time.Until(task.timeout) <= 0 && task.timeout.Unix() != 0 {
				util.Warning.Printf("Sending timeout signal to T[%s]\n", task.taskId)
				task.Cancel()
				task.error(ErrTaskTimeout)
			}
		}
	}
}

func (wp *workerPool) bufferDrainer() {
	for wp.exitFlag == 0 {
		if len(wp.taskBuffer) != 0 && len(wp.taskStream) == 0 {
			wp.taskBufferMu.Lock()
			util.Debug.Println("Draining the buffer!")
			for _, t := range wp.taskBuffer {
				wp.taskStream <- t
			}
			wp.taskBuffer = []*task{}
			wp.taskBufferMu.Unlock()
		}
		time.Sleep(time.Second * 10)
	}

	util.ErrTrace(errors.New("buffer drainer exited"))
}

func (wp *workerPool) addToTaskBuffer(tasks []*task) {
	wp.taskBufferMu.Lock()
	defer wp.taskBufferMu.Unlock()
	util.Debug.Printf("Adding %d tasks to the buffer!", len(tasks))
	wp.taskBuffer = append(wp.taskBuffer, tasks...)
}

// Run launches the standard threads for this worker pool
func (wp *workerPool) Run() {
	wp.exitFlag = 0
	// Spawn the timeout checker
	go wp.reaper()

	// Spawn the buffer worker
	go wp.bufferDrainer()

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

func (wp *workerPool) removeTask(taskKey types.TaskId) {
	wp.taskMu.Lock()
	defer wp.taskMu.Unlock()
	delete(wp.taskMap, taskKey)
}

// For debugging only. This does nothing if the DEV_MODE env var is set to false, as no stopwatch results
// will show then
const printStopwatchResults = false

// Main worker method, spawn a worker and loop over the task channel
//
// `replacement` specifies if the worker is a temporary replacement for another
// task that is parking for a long time. Replacement workers behave a bit
// different to minimize parked time of the other task.
func (wp *workerPool) execWorker(replacement bool) {
	wp.currentWorkers.Add(1)
	go func(workerId int64) {

		// Inc alive workers
		defer wp.currentWorkers.Add(-1)
		util.Debug.Printf("Worker %d reporting for duty o7", workerId)

	WorkLoop:
		for t := range wp.taskStream {

			// Even if we get a new task, if the wp is marked for exit, we just exit
			if wp.exitFlag == 1 {
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
				wp.addToTaskBuffer(tBuf)
			}

			// Inc tasks being processed
			wp.busyCount.Add(1)
			t.SwLap("Task start")
			safetyWork(t, workerId)
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
			t.caster.PushTaskUpdate(t.taskId, TaskComplete,
				types.TaskResult{"task_id": t.taskId, "exit_status": t.exitStatus})
			t.caster.UnsubTask(t)

			// Potentially find the task pool that houses this task pool. All child
			// task pools report their status to the root task pool as well.
			// Do not use any global pool as the root
			rootTaskPool := t.taskPool.GetRootPool().(*taskPool)

			if printStopwatchResults {
				t.sw.PrintResults(true)
			} else if t.exitStatus != TaskSuccess {
				if !rootTaskPool.IsGlobal() {
					rootTaskPool.AddError(t)
				}
				util.Warning.Printf("T[%s] exited with non-success status: %s\n", t.taskId, t.exitStatus)
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
				util.Debug.Printf("Uncompleted tasks on tp created by %s: %d",
					directParent.CreatedInTask().TaskId(), uncompletedTasks-1)
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

func (wp *workerPool) NewVirtualTaskPool() types.TaskPool {
	newQueue := &taskPool{
		totalTasks:     &atomic.Int64{},
		completedTasks: &atomic.Int64{},
		waiterCount:    &atomic.Int32{},
		waiterGate:     &sync.Mutex{},
		exitLock:       &sync.Mutex{},
		workerPool:     wp,
	}

	// The waiterGate begins locked, and will only unlock when all
	// tasks have been queued and finish, then the waiters are released
	newQueue.waiterGate.Lock()

	return newQueue
}

// Status returns the count of tasks in the queue, the total number of tasks accepted,
// number of busy workers, and the total number of live workers in the worker pool
func (wp *workerPool) Status() (int, int, int, int) {
	total := int(wp.lifetimeQueuedCount.Load())

	return len(wp.taskStream), total, int(wp.busyCount.Load()), int(wp.currentWorkers.Load())
}

func (wp *workerPool) statusReporter() {
	for {
		time.Sleep(time.Second * 10)
		remaining, total, busy, alive := wp.Status()
		if busy != 0 {
			util.Info.Printf("Task worker pool status (queued/total, #busy, #alive): %d/%d, %d, %d",
				remaining, total, busy, alive)
		}
	}
}
