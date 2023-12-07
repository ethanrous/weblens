package dataProcess

import (
	"runtime/debug"
	"sync/atomic"

	"github.com/ethrousseau/weblens/api/util"
)

type taskWrapper struct {
	virtualQueueId string
	work func()
}
type workChannel chan taskWrapper

type virtualTaskPool struct {
	queueId string
	totalTasks *atomic.Int64
	completedTasks *atomic.Int64
	waitersCount *atomic.Int64
	waiterChan chan bool
	allQueuedFlag bool
}

type WorkerPool struct {
	maxWorker int
	currentWorkers *atomic.Int64
	taskStream workChannel
	virtualQueues map[string]virtualTaskPool
	busyCount *atomic.Int64

	exitFlag int
}

func NewWorkerPool(workerCount int) (WorkerPool) {
	if workerCount == 0 {
		workerCount = 1
	}
	util.Info.Printf("Starting new worker pool with %d workers", workerCount)

	var busyCount atomic.Int64
	var startingWorkers atomic.Int64

	newWp := WorkerPool{
		maxWorker: workerCount,
		currentWorkers: &startingWorkers,
		taskStream: make(workChannel, workerCount * 100),
		virtualQueues: make(map[string]virtualTaskPool),
		busyCount: &busyCount,
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

func (wp *WorkerPool) workerRecover(workerId int) {
	err := recover()
	wp.currentWorkers.Add(-1)
	if err != nil {
		wp.busyCount.Add(-1)
		util.Error.Printf("Worker %d recovered error: %s\n%s", workerId, err, debug.Stack())
	}
	if wp.exitFlag == 1 {
		util.Debug.Println("Worker ", workerId, " exiting")
		return
	}
	util.Debug.Println("Worker ", workerId, " restarting")
	wp.AddWorker(workerId)
}

func (wp *WorkerPool) AddWorker(workerId int) {
	go func(workerId int) {

		defer wp.workerRecover(workerId)

		wp.currentWorkers.Add(1)
		for task := range wp.taskStream {
			if wp.exitFlag == 1 {break}

			// util.Debug.Println("Worker ", workerId, "starting task")
			wp.busyCount.Add(1)
			task.work()
			wp.busyCount.Add(-1)
			// util.Debug.Println("Worker ", workerId, "finished task")

			wp.virtualQueues[task.virtualQueueId].completedTasks.Add(1)

			if task.virtualQueueId != "GLOBAL" {
				uncompletedTasks := wp.virtualQueues[task.virtualQueueId].totalTasks.Load() - wp.virtualQueues[task.virtualQueueId].completedTasks.Load()
				if uncompletedTasks == 0 && wp.virtualQueues[task.virtualQueueId].allQueuedFlag {
					for wp.virtualQueues[task.virtualQueueId].waitersCount.Load() != 0 {
						wp.virtualQueues[task.virtualQueueId].waiterChan <- true
					}
				}
			}
		}
	}(workerId)
}

func (wp *WorkerPool) AddTask(f func(), queueKey string) {
	t := taskWrapper{
		virtualQueueId: queueKey,
		work: f,
	}

	wp.virtualQueues[queueKey].totalTasks.Add(1)
	wp.taskStream <- t
}

func (wp *WorkerPool) NewVirtualTaskQueue(queueKey string) {
	var totalTasks atomic.Int64
	var completedTasks atomic.Int64
	var waitersCount atomic.Int64

	newQueue := virtualTaskPool{
		queueId: queueKey,
		totalTasks: &totalTasks,
		completedTasks: &completedTasks,
		waitersCount: &waitersCount,
		waiterChan: make(chan bool, wp.currentWorkers.Load()),
	}
	wp.virtualQueues[queueKey] = newQueue
}

func (wp *WorkerPool) NotifyAllQueued(queueKey string) {
	q := wp.virtualQueues[queueKey]
	q.allQueuedFlag = true
	wp.virtualQueues[queueKey] = q
}

func (wp *WorkerPool) CloseWorkQueue(queueKey string) {
	delete(wp.virtualQueues, queueKey)
}

// Returns the count of tasks in the queue, and the total number of tasks accepted, number of busy workers, the total number of live workers in the worker pool
func (wp *WorkerPool) Status(queueKey string) (int, int, int, int) {
	return len(wp.taskStream), int(wp.virtualQueues[queueKey].totalTasks.Load()), int(wp.busyCount.Load()), int(wp.currentWorkers.Load())
}

func (wp *WorkerPool) Close() {
	wp.exitFlag = 1
}

func (wp *WorkerPool) Wait(queueKey string) {
	wp.virtualQueues[queueKey].waitersCount.Add(1)
	<-wp.virtualQueues[queueKey].waiterChan
	wp.virtualQueues[queueKey].waitersCount.Add(-1)
}