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

func NewWorkerPool(initWorkers int) (*WorkerPool, *taskPool) {
	if initWorkers == 0 {
		initWorkers = 1
	}
	util.Info.Printf("Starting new worker pool with %d workers", initWorkers)

	newWp := &WorkerPool{
		maxWorkers:          &atomic.Int64{},
		currentWorkers:      &atomic.Int64{},
		busyCount:           &atomic.Int64{},
		lifetimeQueuedCount: &atomic.Int64{},
		taskStream:          make(workChannel, initWorkers*1000),
		taskBuffer:          []*task{},
		taskBufferMu:        &sync.Mutex{},
		hitStream:           make(hitChannel, initWorkers*2),

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

// The reaper handles timed cancellation of tasks. If a task might
// wait on some action from the client, it should queue a timeout with
// the reaper so it doesn't hang forever
func (wp *WorkerPool) reaper() {
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
			if !task.completed && time.Until(task.timeout) <= 0 && task.timeout.Unix() != 0 {
				task.error(ErrTaskTimeout)
				task.Cancel()
			}
		}
	}
}

func (wp *WorkerPool) bufferDrainer() {
	for wp.exitFlag == 0 {
		if len(wp.taskBuffer) != 0 && len(wp.taskStream) == 0 {
			wp.taskBufferMu.Lock()
			for _, t := range wp.taskBuffer {
				wp.taskStream <- t
			}
			wp.taskBuffer = []*task{}
			wp.taskBufferMu.Unlock()
		}
		time.Sleep(time.Second * 10)
	}
}

func (wp *WorkerPool) addToTaskBuffer(tasks []*task) {
	wp.taskBufferMu.Lock()
	wp.taskBuffer = append(wp.taskBuffer, tasks...)
	wp.taskBufferMu.Unlock()
}

// Launch the standard threads for this worker pool
func (wp *WorkerPool) Run() {
	wp.exitFlag = 0
	// Spawn the timeout checker
	go wp.reaper()

	// Spawn the buffer worker
	go wp.bufferDrainer()

	var i int64
	for i = 0; i < wp.maxWorkers.Load(); i++ {
		// These are the base, 'omnipresent' threads for this pool,
		// so they are NOT replacement workers. See wp.execWorker
		wp.execWorker(false)
	}
}

// For debugging only. This does nothing if the DEV_MODE env var is set to false, as no stopwatch results will show then.
const printStopwatchResults = false

// Main worker method, spawn a worker and loop over the task channel
//
// `replacement` specifies if the worker is a temporary replacement for another
// task that is parking for a long time. Replacement workers behave a bit
// different to minimize parked time of the other task.
func (wp *WorkerPool) execWorker(replacement bool) {
	go func(workerId int64) {

		// Inc alive workers
		wp.currentWorkers.Add(1)

	WorkLoop:
		for t := range wp.taskStream {

			// Even if we get a new task, if the wp is marked for exit, we just exit
			if wp.exitFlag == 1 {
				// We don't care about the exitLock here, since the whole wp
				// is going down anyway.

				// Dec alive workers
				wp.currentWorkers.Add(-1)

				break WorkLoop
			}

			// Replacement workers are not allowed to do "scan_directory" tasks
			if replacement && t.taskType == ScanDirectoryTask {

				// If there are twice the number of free spaces in the chan, don't bother pulling
				// everything into the waiting buffer, just put it at the end right now.
				if cap(wp.taskStream)-len(wp.taskStream) > int(wp.currentWorkers.Load())*2 {
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
			if !t.completed {
				t.success("closed by worker pool")
			}

			if printStopwatchResults {
				t.sw.PrintResults(true)
			} else if t.exitStatus != TaskSuccess {
				util.Warning.Printf("T[%s] exited with non-success status: %s\n", t.taskId, t.exitStatus)
			}

			// Potentially find the task pool that houses this task pool. All child
			// task pools report their status to the root task pool as well.
			// Do not use any global pool as the root
			rootTaskPool := t.taskPool
			for rootTaskPool.parentTaskPool != nil && !rootTaskPool.parentTaskPool.treatAsGlobal {
				rootTaskPool = rootTaskPool.parentTaskPool
			}

			if !t.persistent {
				removeTask(t.taskId)
			}

			canContinue := true
			if rootTaskPool == t.taskPool {
				// Updating the number of workers and then checking it's value is dangerous
				// to do concurrently. Specifically, the waiterGate lock on the queue will,
				// very rarely, attempt to unlock twice if two tasks finish at the same time.
				// So we must treat this whole area as a critical section
				t.taskPool.exitLock.Lock()

				// Set values and notifications now that task has completed. Returns
				// a bool that specifies if this thread should continue and grab another
				// task, or if it should exit
				canContinue = t.taskPool.handleTaskExit(replacement)

				t.taskPool.exitLock.Unlock()
			} else {
				// Must hold both locks (and must acquire root lock first) to enter a dual-update.
				// Any other ordering will result in race conditions or deadlocks
				rootTaskPool.exitLock.Lock()

				t.taskPool.exitLock.Lock()
				canContinue = t.taskPool.handleTaskExit(replacement)

				// We *should* get the same canContinue value from here, so we do not
				// check it a second time. If we *don't* get the same value, we can safely ignore it
				rootTaskPool.handleTaskExit(replacement)

				t.taskPool.exitLock.Unlock()
				rootTaskPool.exitLock.Unlock()
			}

			if !canContinue {
				break WorkLoop
			}

		}

		// If currentWorkers is 4, we then have worker ids from 0-3,
		// so currentWorkers becomes the new worker id as it's next
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
func (wp *WorkerPool) addReplacementWorker() {
	wp.maxWorkers.Add(1)
	wp.execWorker(true)
}

func (wp *WorkerPool) removeWorker() {
	wp.maxWorkers.Add(-1)
}

func (wp *WorkerPool) NewVirtualTaskPool() *taskPool {
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

func (tp *taskPool) Cancel() {
	// TODO
}

func (tp *taskPool) QueueTask(Task types.Task) (err error) {
	t := Task.(*task)
	if tp.workerPool.exitFlag == 1 {
		util.Warning.Println("Not queuing task while worker pool is going down")
		return
	}

	if t.err != nil {
		// Tasks that have failed will not be re-tried. If the errored task is removed from the
		// task map, then it will be re-tried because the previous error was lost. This can be
		// sometimes be useful, some tasks auto-remove themselves after they finish.
		util.Warning.Println("Not re-queueing task that has error set, please restart weblens to try again")
		return
	}

	if t.taskPool != nil {
		// Task is already queued, we are not allowed to move it to another queue.
		// We can call .ClearAndRecompute() on the task and it will queue it
		// again, but it cannot be transferred
		if t.taskPool != tp {
			util.Warning.Println("Attempted to re-assign task to another queue")
		}
		return
	}

	if tp.allQueuedFlag {
		// We cannot add tasks to a queue that has been closed
		return errors.New("attempting to add task to closed task queue")
	}

	tp.totalTasks.Add(1)

	if tp.parentTaskPool != nil {
		tmpTp := tp
		for tmpTp.parentTaskPool != nil {
			tmpTp = tmpTp.parentTaskPool
		}
		tmpTp.totalTasks.Add(1)
	}

	// Set the tasks queue
	t.taskPool = tp

	tp.workerPool.lifetimeQueuedCount.Add(1)

	// Put the task in the queue
	tp.workerPool.taskStream <- t

	return
}

// Specify the work queue as being a "global" one
func (tp *taskPool) MarkGlobal() {
	tp.treatAsGlobal = true
}

func (tp *taskPool) SignalAllQueued() {
	if tp.treatAsGlobal {
		util.Error.Println("Attempt to signal all queued for global queue")
	}

	tp.exitLock.Lock()
	// If all tasks finish (e.g. early failure, etc.) before we signal that they are all queued,
	// the final exiting task will not let the waiters out, so we must do it here
	if tp.completedTasks.Load() == tp.totalTasks.Load() {
		tp.waiterGate.Unlock()
	}
	tp.allQueuedFlag = true
	tp.exitLock.Unlock()

	if tp.hasQueueThread {
		tp.workerPool.removeWorker()
		tp.hasQueueThread = false
	}
}

// Returns the count of tasks in the queue, the total number of tasks accepted,
// number of busy workers, and the total number of live workers in the worker pool
func (wp *WorkerPool) Status() (int, int, int, int) {
	total := int(wp.lifetimeQueuedCount.Load())

	return len(wp.taskStream), total, int(wp.busyCount.Load()), int(wp.currentWorkers.Load())
}
