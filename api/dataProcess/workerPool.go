package dataProcess

import (
	"errors"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

type workChannel chan *task

type virtualTaskPool struct {
	treatAsGlobal bool
	totalTasks *atomic.Int64
	completedTasks *atomic.Int64
	waiterCount *atomic.Int32
	waitMu *sync.Mutex
	allQueuedFlag bool
	parentWorkerPool *WorkerPool
}

type WorkerPool struct {
	maxWorker int
	currentWorkers *atomic.Int64
	taskStream workChannel
	busyCount *atomic.Int64

	lifetimeQueuedCount *atomic.Int64
	exitFlag int
}

func NewWorkerPool(workerCount int) (WorkerPool, *virtualTaskPool) {
	if workerCount == 0 {
		workerCount = 1
	}
	util.Info.Printf("Starting new worker pool with %d workers", workerCount)

	var busyCount atomic.Int64
	var startingWorkers atomic.Int64
	var totalTasks atomic.Int64

	newWp := WorkerPool{
		maxWorker: workerCount,
		currentWorkers: &startingWorkers,
		taskStream: make(workChannel, workerCount * 1000),
		busyCount: &busyCount,
		lifetimeQueuedCount: &totalTasks,
		exitFlag: 0,
	}

	globalPool := newWp.NewVirtualTaskQueue()
	globalPool.MarkGlobal()

	return newWp, globalPool
}

func workerRecover(task *task, workerId int64) {
	err := recover()
	if err != nil {
		task.err = err
		util.Error.Printf("Worker %d recovered error: %s\n%s", workerId, err, debug.Stack())
	}
}

func saftyWork(task *task, workerId int64) {
	defer workerRecover(task, workerId)
	task.work()
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		wp.AddWorker()
	}
}

func (wp *WorkerPool) AddWorker() {
	go func(workerId int64) {
		wp.currentWorkers.Add(1)
		for task := range wp.taskStream {
			if wp.exitFlag == 1 {break}

			wp.busyCount.Add(1)
			saftyWork(task, workerId)
			wp.busyCount.Add(-1)

			task.queue.completedTasks.Add(1)

			task.waitMu.Unlock()

			if !task.queue.treatAsGlobal {
				uncompletedTasks := task.queue.totalTasks.Load() - task.queue.completedTasks.Load()
				if uncompletedTasks == 0 && task.queue.allQueuedFlag {
					task.queue.waitMu.Unlock()
					// Make sure all waiters have left before closing the queue, spin and sleep for 10ms if not
					// Should only loop once, just needs to wait for the waiters to lock and release. 10ms is an eternity for that
					for task.queue.waiterCount.Load() != 0 {time.Sleep(10000000)}
				}
			} else if wp.currentWorkers.Load() > int64(wp.maxWorker) {
				break
			}
		}
		wp.currentWorkers.Add(-1)
	}(wp.currentWorkers.Load())
}

func (wp *WorkerPool) NewVirtualTaskQueue() *virtualTaskPool {
	var totalTasks atomic.Int64
	var completedTasks atomic.Int64
	var waiterCount atomic.Int32

	newQueue := &virtualTaskPool{
		totalTasks: &totalTasks,
		completedTasks: &completedTasks,
		waiterCount: &waiterCount,
		waitMu: &sync.Mutex{},
	}
	newQueue.waitMu.Lock()

	newQueue.parentWorkerPool = wp

	return newQueue
}

func (wp *WorkerPool) Close() {
	wp.exitFlag = 1
}

func (wq *virtualTaskPool) QueueTask(t *task) {
	if t.queue != nil {
		// Task is already queued, so we don't need to re-add to the queue
		if (t.queue != wq) {
			util.Warning.Println("Attempted to re-assign task to another queue")
		}
		return
	}

	if wq.allQueuedFlag {
		panic(errors.New("attempting to add task to closed task queue"))
	}

	if wq.treatAsGlobal {
		wq.parentWorkerPool.AddWorker()
	}
	wq.totalTasks.Add(1)
	t.queue = wq

	wq.parentWorkerPool.lifetimeQueuedCount.Add(1)
	wq.parentWorkerPool.taskStream <- t
}

func (wq *virtualTaskPool) MarkGlobal() {
	wq.treatAsGlobal = true
}

func (wq *virtualTaskPool) AllQueued() {
	wq.allQueuedFlag = true
}

// Returns the count of tasks in the queue, and the total number of tasks accepted, number of busy workers, the total number of live workers in the worker pool
func (wq *virtualTaskPool) Status() (int, int, int, int) {
	var total int
	if wq.treatAsGlobal {
		total = int(wq.parentWorkerPool.lifetimeQueuedCount.Load())
	} else {
		total = int(wq.totalTasks.Load())
	}

	return len(wq.parentWorkerPool.taskStream), total, int(wq.parentWorkerPool.busyCount.Load()), int(wq.parentWorkerPool.currentWorkers.Load())
}

func (wq *virtualTaskPool) Wait() {
	if wq.allQueuedFlag && wq.totalTasks.Load() == 0 {
		return
	}

	wq.waiterCount.Add(1)
	util.Debug.Println("Parking")
	wq.waitMu.Lock()
	//lint:ignore SA2001 We want to wake up when the task is finished, and then signal other waiters to do the same
	wq.waitMu.Unlock()
	util.Debug.Println("Woke")
	wq.waiterCount.Add(-1)
}