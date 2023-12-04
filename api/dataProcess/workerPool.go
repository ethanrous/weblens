package dataProcess

import (
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/api/util"
)

type task struct {
	work func()
	flag int
}

type WorkerPool struct {
	maxWorker int
	currentWorkers *atomic.Int64
	remainingTasks chan task
	busyCount *atomic.Int64
	totalTasks *atomic.Int64
	exitFlag int
}

func NewWorkerPool(workerCount int) (WorkerPool) {
	if workerCount == 0 {
		workerCount = 1
	}
	util.Info.Printf("Starting new worker pool with %d workers", workerCount)
	var busyCount atomic.Int64
	var totalTasks atomic.Int64
	var startingWorkers atomic.Int64
	newWp := WorkerPool{
		maxWorker: workerCount,
		currentWorkers: &startingWorkers,
		remainingTasks: make(chan task, workerCount * 1000),
		busyCount: &busyCount,
		totalTasks: &totalTasks,
		exitFlag: 0,
	}
	return newWp
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		wp.AddWorker(i)
	}
}

func (wp *WorkerPool) AddWorker(workerId int) {
	go func(workerID int) {
		defer func(){
			err := recover()
			wp.currentWorkers.Add(-1)
			if err != nil {
				wp.busyCount.Add(-1)
				util.Error.Printf("Worker %d recovered error: %s\n%s", workerID, err, debug.Stack())
				return
			}
			util.Debug.Println("Worker ", workerID, " restarting")
			wp.AddWorker(workerID)
		}()
		wp.currentWorkers.Add(1)
		for task := range wp.remainingTasks {
			if task.flag == 1 {break}
			wp.busyCount.Add(1)
			task.work()
			wp.busyCount.Add(-1)
		}
	}(workerId)
}

func (wp *WorkerPool) AddTask(f func()) {
	wp.totalTasks.Add(1)
	wp.remainingTasks <- task{work: f, flag: 0}
}

// Returns the count of tasks in the queue, and the total number of tasks accepted, number of busy workers, the total number of live workers in the worker pool
func (wp *WorkerPool) Status() (int, int, int, int) {
	return len(wp.remainingTasks), int(wp.totalTasks.Load()), int(wp.busyCount.Load()), int(wp.currentWorkers.Load())
}

func (wp *WorkerPool) Close() {
	for wp.currentWorkers.Load() > 0 {
		wp.remainingTasks <- task{flag: 1}
	}
}

func (wp *WorkerPool) Wait() {
	for wp.busyCount.Load() > 0 {
		time.Sleep(time.Second)
	}
}