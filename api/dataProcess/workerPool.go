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

type workChannel chan *Task
type hit struct {
	time   time.Time
	target *Task
}
type hitChannel chan hit

type virtualTaskPool struct {
	treatAsGlobal    bool
	totalTasks       *atomic.Int64
	completedTasks   *atomic.Int64
	waiterCount      *atomic.Int32
	waiterGate       *sync.Mutex
	exitLock         *sync.Mutex
	allQueuedFlag    bool
	parentWorkerPool *WorkerPool
}

type WorkerPool struct {
	maxWorkers     *atomic.Int64 // Max allowed worker count
	currentWorkers *atomic.Int64 // Currnet worker count
	busyCount      *atomic.Int64 // Number of workers currently executing a task

	lifetimeQueuedCount *atomic.Int64

	taskStream workChannel
	hitStream  hitChannel

	exitFlag int
}

func NewWorkerPool(initWorkers int) (*WorkerPool, *virtualTaskPool) {
	if initWorkers == 0 {
		initWorkers = 1
	}
	util.Info.Printf("Starting new worker pool with %d workers", initWorkers)

	var maxWorkers atomic.Int64
	var busyCount atomic.Int64
	var startingWorkers atomic.Int64
	var totalTasks atomic.Int64

	newWp := &WorkerPool{
		maxWorkers:          &maxWorkers,
		currentWorkers:      &startingWorkers,
		taskStream:          make(workChannel, initWorkers*1000),
		hitStream:           make(hitChannel, initWorkers*2),
		busyCount:           &busyCount,
		lifetimeQueuedCount: &totalTasks,
		exitFlag:            1,
	}

	newWp.maxWorkers.Add(int64(initWorkers))

	globalPool := newWp.NewVirtualTaskQueue()
	globalPool.MarkGlobal()

	return newWp, globalPool
}

func workerRecover(task *Task, workerId int64) {
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

func saftyWork(task *Task, workerId int64) {
	defer workerRecover(task, workerId)
	task.work()
}

// The executioner handles timed cancellation of tasks. If a task might
// wait on some action from the client, it should queue a timeout with
// the executioner so it doesn't hang forever
func (wp *WorkerPool) executioner() {
	timerStream := make(chan *Task)
	for {
		select {
		case newHit := <-wp.hitStream:
			go func(h hit) { time.Sleep(time.Until(h.time)); timerStream <- h.target }(newHit)
		case task := <-timerStream:
			// Possible that the task has kicked it's timeout down the road
			// since it first queued it, we must check that we are past the
			// timeout before cancelling the task. Also check that it
			// has not already finished
			if time.Until(task.timeout) <= 0 && !task.completed {
				task.Cancel()
				err := errors.New("timeout")
				task.error(err)
			}
		}
	}
}

func (wp *WorkerPool) Run() {
	wp.exitFlag = 0
	go wp.executioner()

	var i int64
	for i = 0; i < wp.maxWorkers.Load(); i++ {
		wp.execWorker(false)
	}
}

// Main worker method, spawn a worker and loop over the task channel
//
// `replacement` defines if the worker is a temporary replacement for another
// task that is parking for a long time. Replacement workers have a more limited
// allowed ruleset to minimize parked time of the other task.
// They and will be removed some time after the main task wakes up.
func (wp *WorkerPool) execWorker(replacement bool) {
	go func(workerId int64) {

		// Inc alive workers
		wp.currentWorkers.Add(1)

	WorkLoop:
		for task := range wp.taskStream {
			if wp.exitFlag == 1 {
				// We don't care about the exitLock here, since the whole wp
				// is going down anyway.

				// Dec alive workers
				wp.currentWorkers.Add(-1)

				break WorkLoop
			}
			if replacement && task.taskType == "scan_directory" {
				wp.taskStream <- task
				continue WorkLoop
			}

			// Inc tasks being processed
			wp.busyCount.Add(1)
			saftyWork(task, workerId)
			// Dec tasks being processed
			wp.busyCount.Add(-1)

			task.queue.completedTasks.Add(1)

			// Wake any waiters on this task
			task.waitMu.Unlock()

			// Tasks must set their completed flag before exiting
			// if it wasn't done in the work body, we do it for them
			if !task.completed {
				task.success("closed by worker pool")
			}

			// Updating the number of workers and then checking it's value is dangerous
			// to do concurrently. Specifically, the waiterGate lock on the queue will,
			// very rarely, attempt to unlock twice if two tasks finish at the same time.
			// So we must treat this whole area as a critical section
			task.queue.exitLock.Lock()

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
			if replacement && wp.currentWorkers.Load() > wp.maxWorkers.Load() {
				// Dec alive workers
				wp.currentWorkers.Add(-1)

				// Important to decrement alive workers inside the exit lock, so
				// we don't have multiple workers exiting when we only need the 1
				task.queue.exitLock.Unlock()
				/*
					Tasks can create "replacement" workers as to not create a deadlock.
					e.g. all n `wp.maxWorkers` are parked "scan_directory" tasks waiting
					for their "scan_file" children tasks to complete, but never will since
					all worker threads are taken up by blocked tasks.

					Replacement workers:
					1. Are not allowed to execute "scan_directory" tasks
					2. will exit once the main thread has woken up, and shrunk the worker pool back down

					So, if this is a replacement task, and we have more workers than the pool allows, we exit
				*/
				break WorkLoop
			}

			// If we have already began running the task,
			// we must finish and clean up before checking exitFlag again here.
			// the task *could* be cancelled to speed things up, but that
			// is not our job.
			if wp.exitFlag == 1 {
				// Dec alive workers
				wp.currentWorkers.Add(-1)
				task.queue.exitLock.Unlock()
				break WorkLoop
			}

			task.queue.exitLock.Unlock()

		}

		// If the number of alive workers is 4, we have worker ids from 0-3,
		// so the number of workers (4, in this example) becomes the new worker id
	}(wp.currentWorkers.Load())
}

func (wp *WorkerPool) addReplacementWorker() {
	wp.maxWorkers.Add(1)
	wp.execWorker(true)
}

func (wp *WorkerPool) removeWorker() {
	wp.maxWorkers.Add(-1)
}

func (wp *WorkerPool) NewVirtualTaskQueue() *virtualTaskPool {
	var totalTasks atomic.Int64
	var completedTasks atomic.Int64
	var waiterCount atomic.Int32

	newQueue := &virtualTaskPool{
		totalTasks:     &totalTasks,
		completedTasks: &completedTasks,
		waiterCount:    &waiterCount,
		waiterGate:     &sync.Mutex{},
		exitLock:       &sync.Mutex{},
	}
	newQueue.waiterGate.Lock()

	newQueue.parentWorkerPool = wp

	return newQueue
}

func (wq *virtualTaskPool) Cancel() {

}

func (wq *virtualTaskPool) QueueTask(t *Task) {
	if wq.parentWorkerPool.exitFlag == 1 {
		return
	}

	if t.err != nil {
		util.Warning.Println("Not re-queueing task that has error set, please restart weblens to try again")
		return
	}

	if t.queue != nil {
		// Task is already queued, so we don't need to re-add to the queue
		if t.queue != wq {
			util.Warning.Println("Attempted to re-assign task to another queue")
		}
		return
	}

	if wq.allQueuedFlag {
		panic(errors.New("attempting to add task to closed task queue"))
	}

	wq.totalTasks.Add(1)
	t.queue = wq

	wq.parentWorkerPool.lifetimeQueuedCount.Add(1)
	wq.parentWorkerPool.taskStream <- t
}

func (wq *virtualTaskPool) MarkGlobal() {
	wq.treatAsGlobal = true
}

func (wq *virtualTaskPool) SignalAllQueued() {
	if wq.treatAsGlobal {
		util.Error.Println("Attempt to signal all queued for global queue")
	}

	wq.allQueuedFlag = true
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
