package dataProcess

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

func NewWorkerPool(initWorkers int) (*WorkerPool, *virtualTaskPool) {
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
		hitStream:           make(hitChannel, initWorkers*2),

		// Worker pool starts disabled
		exitFlag: 1,
	}

	newWp.maxWorkers.Add(int64(initWorkers))

	// Worker pool always has one global queue
	globalPool := newWp.NewVirtualTaskQueue()
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
		util.Error.Printf("Worker %d recovered error: %s\n%s", workerId, err, debug.Stack())
		task.error(err.(error))
	}
}

func saftyWork(task *task, workerId int64) {
	defer workerRecover(task, workerId)
	task.work()
}

// The executioner handles timed cancellation of tasks. If a task might
// wait on some action from the client, it should queue a timeout with
// the executioner so it doesn't hang forever
func (wp *WorkerPool) executioner() {
	timerStream := make(chan *task)
	for {
		select {
		case newHit := <-wp.hitStream:
			go func(h hit) { time.Sleep(time.Until(h.time)); timerStream <- h.target }(newHit)
		case task := <-timerStream:
			// Possible that the task has kicked it's timeout down the road
			// since it first queued it, we must check that we are past the
			// timeout before cancelling the task. Also check that it
			// has not already finished
			if !task.completed && time.Until(task.timeout) <= 0 && task.timeout.Unix() != 0 {
				err := errors.New("timeout")
				task.error(err)
				task.Cancel()
			}
		}
	}
}

// Launch the standard threads for this worker pool
func (wp *WorkerPool) Run() {
	wp.exitFlag = 0
	// Spawn the timeout checker
	go wp.executioner()

	var i int64
	for i = 0; i < wp.maxWorkers.Load(); i++ {
		// These are the base, everpresent threads for this pool,
		// so they are NOT replacement workers. See wp.execWorker
		wp.execWorker(false)
	}
}

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
		for task := range wp.taskStream {

			// Even if we get a new task, if the wp is marked for exit, we just exit
			if wp.exitFlag == 1 {
				// We don't care about the exitLock here, since the whole wp
				// is going down anyway.

				// Dec alive workers
				wp.currentWorkers.Add(-1)

				break WorkLoop
			}
			// Replacement workers are not allowed to do "scan_directory" tasks
			if replacement && task.taskType == string(ScanDirectoryTask) {
				wp.taskStream <- task
				continue WorkLoop
			}

			// Inc tasks being processed
			wp.busyCount.Add(1)
			task.SwLap("Task start")
			saftyWork(task, workerId)
			// Dec tasks being processed
			wp.busyCount.Add(-1)

			// Wake any waiters on this task
			task.waitMu.Unlock()

			// Tasks must set their completed flag before exiting
			// if it wasn't done in the work body, we do it for them
			if !task.completed {
				task.success("closed by worker pool")
			}

			if task.exitStatus != "cancelled" {
				task.sw.PrintResults()
			}

			// Updating the number of workers and then checking it's value is dangerous
			// to do concurrently. Specifically, the waiterGate lock on the queue will,
			// very rarely, attempt to unlock twice if two tasks finish at the same time.
			// So we must treat this whole area as a critical section
			task.queue.exitLock.Lock()

			task.queue.completedTasks.Add(1)

			// Global queues do not finish and cannot be waited on. If this is NOT a global queue,
			// we check if we are empty and finished, and if so wake any waiters.
			if !task.queue.treatAsGlobal {
				uncompletedTasks := task.queue.totalTasks.Load() - task.queue.completedTasks.Load()

				// Even if we are out of tasks, if we have not been told all tasks
				// were queued, we do not wake the waiters
				if uncompletedTasks == 0 && task.queue.waiterCount.Load() != 0 && task.queue.allQueuedFlag {
					task.queue.waiterGate.Unlock()

					// Check if all waiters have awoken before closing the queue, spin and sleep for 10ms if not
					// Should only loop a handful of times, if at all, we just need to wait for the waiters to
					// lock and then release immediately, should take nanoseconds each
					for task.queue.waiterCount.Load() != 0 {
						time.Sleep(10000000)
					}
				}
			}
			// If this is a replacement task, and we have more workers than the target for the pool, we exit
			if replacement && wp.currentWorkers.Load() > wp.maxWorkers.Load() {
				// Important to decrement alive workers inside the exitLock, so
				// we don't have multiple workers exiting when we only need the 1
				wp.currentWorkers.Add(-1)

				task.queue.exitLock.Unlock()
				break WorkLoop
			}

			// If we have already began running the task,
			// we must finish and clean up before checking exitFlag again here.
			// The task *could* be cancelled to speed things up, but that
			// is not our job.
			if wp.exitFlag == 1 {
				// Dec alive workers
				wp.currentWorkers.Add(-1)
				task.queue.exitLock.Unlock()
				break WorkLoop
			}

			task.queue.exitLock.Unlock()

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

func (wp *WorkerPool) NewVirtualTaskQueue() *virtualTaskPool {
	newQueue := &virtualTaskPool{
		totalTasks:       &atomic.Int64{},
		completedTasks:   &atomic.Int64{},
		waiterCount:      &atomic.Int32{},
		waiterGate:       &sync.Mutex{},
		exitLock:         &sync.Mutex{},
		parentWorkerPool: wp,
	}

	// The waiterGate begins locked, and will only unlock when all
	// tasks have been queued and finish, then the waiters are released
	newQueue.waiterGate.Lock()

	return newQueue
}

func (wq *virtualTaskPool) Cancel() {
	// TODO
}

func (wq *virtualTaskPool) QueueTask(t *task) (err error) {
	if wq.parentWorkerPool.exitFlag == 1 {
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

	if t.queue != nil {
		// Task is already queued, we are not allowed to move it to another queue.
		// We can call .ClearAndRecompute() on the task and it will queue it
		// again, but it cannot be transfered
		if t.queue != wq {
			util.Warning.Println("Attempted to re-assign task to another queue")
		}
		return
	}

	if wq.allQueuedFlag {
		// We cannot add tasks to a queue that has been closed
		return errors.New("attempting to add task to closed task queue")
	}

	wq.totalTasks.Add(1)

	// Set the tasks queue
	t.queue = wq

	wq.parentWorkerPool.lifetimeQueuedCount.Add(1)

	// Put the task in the queue
	wq.parentWorkerPool.taskStream <- t

	return
}

// Specifcy the work queue as being a "global" one
func (wq *virtualTaskPool) MarkGlobal() {
	wq.treatAsGlobal = true
}

func (wq *virtualTaskPool) SignalAllQueued() {
	if wq.treatAsGlobal {
		util.Error.Println("Attempt to signal all queued for global queue")
	}

	wq.exitLock.Lock()
	// If all tasks finish (e.g. early failure, etc.) before we signal that they are all queued,
	// the final exiting task will not let the waiters out, so we must do it here
	if wq.completedTasks.Load() == wq.totalTasks.Load() {
		wq.waiterGate.Unlock()
	}
	wq.allQueuedFlag = true
	wq.exitLock.Unlock()
}

func (wq *virtualTaskPool) ClearAllQueued() {
	if wq.waiterCount.Load() != 0 {
		util.Warning.Println("Clearing all queued flag on work queue that still has sleepers")
	}
	wq.allQueuedFlag = false
}

// Returns the count of tasks in the queue, the total number of tasks accepted,
// number of busy workers, and the total number of live workers in the worker pool
func (wq *virtualTaskPool) Status() (int, int, int, int) {
	var total int
	if wq.treatAsGlobal {
		total = int(wq.parentWorkerPool.lifetimeQueuedCount.Load())
	} else {
		total = int(wq.totalTasks.Load())
	}

	return len(wq.parentWorkerPool.taskStream), total, int(wq.parentWorkerPool.busyCount.Load()), int(wq.parentWorkerPool.currentWorkers.Load())
}

// Park the thread on the work queue until all the tasks have been queued and finish.
// **If you never call wq.SignalAllQueued(), the waiters will never wake up**
// Make sure that you SignalAllQueued before parking here if it is the only thread
// loading tasks
func (wq *virtualTaskPool) Wait(supplementWorker bool) {
	// Waiting on global queues does not make sense, they are not meant to end
	// or
	// All the tasks were queued, and they have all finished,
	// so no need to wait, we can "wake up" instantly.
	if wq.treatAsGlobal || (wq.allQueuedFlag && wq.totalTasks.Load() == 0) {
		return
	}

	// If we want to park another thread that is currently executing a task,
	// e.g a directory scan waiting for the child filescans to complete,
	// we want to add an additional worker to the pool temporarily to suppliment this one
	if supplementWorker {
		wq.parentWorkerPool.addReplacementWorker()
	}

	_, file, line, _ := runtime.Caller(1)
	util.Debug.Printf("Parking on worker pool from %s:%d\n", file, line)

	wq.waiterCount.Add(1)
	wq.waiterGate.Lock()
	//lint:ignore SA2001 We want to wake up when the task is finished, and then signal other waiters to do the same
	wq.waiterGate.Unlock()
	util.Debug.Printf("Woke up, returning to %s:%d\n", file, line)
	wq.waiterCount.Add(-1)

	if supplementWorker {
		wq.parentWorkerPool.removeWorker()
	}
}
