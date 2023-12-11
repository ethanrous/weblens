package dataProcess

import (
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

type workChannel chan *task

type virtualTaskPool struct {
	queueId string
	totalTasks *atomic.Int64
	completedTasks *atomic.Int64
	waiterCount *atomic.Int32
	waitMu *sync.Mutex
	allQueuedFlag bool
}

type WorkerPool struct {
	maxWorker int
	currentWorkers *atomic.Int64
	taskStream workChannel
	virtualQueues map[string]virtualTaskPool
	busyCount *atomic.Int64

	virtQueueMu *sync.Mutex
	lifetimeQueuedCount *atomic.Int64
	exitFlag int
}

func NewWorkerPool(workerCount int) (WorkerPool) {
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
		virtualQueues: make(map[string]virtualTaskPool),
		busyCount: &busyCount,
		lifetimeQueuedCount: &totalTasks,
		virtQueueMu: &sync.Mutex{},
		exitFlag: 0,
	}

	newWp.NewVirtualTaskQueue("GLOBAL")

	return newWp
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		wp.AddWorker(i)
	}
}

func workerRecover(task *task, workerId int) {
	err := recover()
	if err != nil {
		task.err = err
		util.Error.Printf("Worker %d recovered error: %s\n%s", workerId, err, debug.Stack())
	}
}

func saftyWork(task *task, workerId int) {
	defer workerRecover(task, workerId)
	task.work()
}

func (wp *WorkerPool) AddWorker(workerId int) {
	go func(workerId int) {
		wp.currentWorkers.Add(1)
		for task := range wp.taskStream {
			if wp.exitFlag == 1 {break}

			wp.busyCount.Add(1)
			saftyWork(task, workerId)
			wp.busyCount.Add(-1)

			wp.virtQueueMu.Lock()
			wp.virtualQueues[task.QueueId].completedTasks.Add(1)
			wp.virtQueueMu.Unlock()
			task.waitMu.Unlock()

			if task.QueueId != "GLOBAL" {
				wp.virtQueueMu.Lock()
				uncompletedTasks := wp.virtualQueues[task.QueueId].totalTasks.Load() - wp.virtualQueues[task.QueueId].completedTasks.Load()
				if uncompletedTasks == 0 && wp.virtualQueues[task.QueueId].allQueuedFlag {
					wp.virtualQueues[task.QueueId].waitMu.Unlock()
					// Make sure all waiters have left before closing the queue, spin and sleep for 10ms if not
					for wp.virtualQueues[task.QueueId].waiterCount.Load() != 0 {time.Sleep(10000000)}
					wp.CloseVirtualQueue(task.QueueId)
				}
				wp.virtQueueMu.Unlock()
			} else if wp.currentWorkers.Load() > int64(wp.maxWorker) {
				break
			}
		}
		wp.currentWorkers.Add(-1)
	}(workerId)
}

func (wp *WorkerPool) AddTask(task *task) {
	if task.QueueId == "GLOBAL" {
		wp.AddWorker(int(wp.currentWorkers.Load()))
	}

	wp.virtQueueMu.Lock()
	wp.virtualQueues[task.QueueId].totalTasks.Add(1)
	wp.virtQueueMu.Unlock()
	wp.lifetimeQueuedCount.Add(1)
	wp.taskStream <- task
}

func (wp *WorkerPool) NewVirtualTaskQueue(queueKey string) {
	var totalTasks atomic.Int64
	var completedTasks atomic.Int64
	var waiterCount atomic.Int32

	newQueue := virtualTaskPool{
		queueId: queueKey,
		totalTasks: &totalTasks,
		completedTasks: &completedTasks,
		waiterCount: &waiterCount,
		waitMu: &sync.Mutex{},
	}
	newQueue.waitMu.Lock()
	wp.virtQueueMu.Lock()
	wp.virtualQueues[queueKey] = newQueue
	wp.virtQueueMu.Unlock()
}

func (wp *WorkerPool) NotifyAllQueued(queueKey string) {
	wp.virtQueueMu.Lock()
	q := wp.virtualQueues[queueKey]
	q.allQueuedFlag = true
	wp.virtualQueues[queueKey] = q
	wp.virtQueueMu.Unlock()
}

func (wp *WorkerPool) CloseVirtualQueue(queueKey string) {
	delete(wp.virtualQueues, queueKey)
}

// Returns the count of tasks in the queue, and the total number of tasks accepted, number of busy workers, the total number of live workers in the worker pool
func (wp *WorkerPool) Status(queueKey string) (int, int, int, int) {
	var total int
	if queueKey == "GLOBAL" {
		total = int(wp.lifetimeQueuedCount.Load())
	} else {
		total = int(wp.virtualQueues[queueKey].totalTasks.Load())
	}
	return len(wp.taskStream), total, int(wp.busyCount.Load()), int(wp.currentWorkers.Load())
}

func (wp *WorkerPool) Close() {
	wp.exitFlag = 1
}

func (wp *WorkerPool) Wait(queueKey string) {
	wp.virtQueueMu.Lock()
	if wp.virtualQueues[queueKey].allQueuedFlag && wp.virtualQueues[queueKey].totalTasks.Load() == 0 {
		wp.virtQueueMu.Unlock()
		return
	}
	wp.virtQueueMu.Unlock()

	wp.virtualQueues[queueKey].waiterCount.Add(1)
	wp.virtualQueues[queueKey].waitMu.Lock()
	//lint:ignore SA2001 We want to wake up when the task is finished, and then signal other waiters to do the same
	wp.virtualQueues[queueKey].waitMu.Unlock()
	wp.virtualQueues[queueKey].waiterCount.Add(-1)
}