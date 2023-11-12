package dataProcess

import (
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
	currentWorkers *int64
	remainingTasks chan task
	busyCount *int64
	totalTasks *int64
}

func NewWorkerPool(workerCount int) (WorkerPool) {
	busyCount := int64(0)
	totalTasks := int64(0)
	startingWorkers := int64(0)
	newWp := WorkerPool{workerCount, &startingWorkers, make(chan task, workerCount * 1000), &busyCount, &totalTasks}
	return newWp
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		go func(workerID int) {
			atomic.StoreInt64(wp.currentWorkers, atomic.AddInt64(wp.currentWorkers, 1))
			for task := range wp.remainingTasks {
				if task.flag == 1 {
					atomic.StoreInt64(wp.currentWorkers, atomic.AddInt64(wp.currentWorkers, -1))
					util.Debug.Println("Worker ", workerID, " Exiting")
					break
				}
				atomic.StoreInt64(wp.busyCount, atomic.AddInt64(wp.busyCount, 1))
				task.work()
				atomic.StoreInt64(wp.busyCount, atomic.AddInt64(wp.busyCount, -1))
			}
		}(i + 1)
	}
}

func (wp *WorkerPool) AddTask(f func()) {
	atomic.StoreInt64(wp.totalTasks, atomic.AddInt64(wp.totalTasks, 1))
	wp.remainingTasks <- task{work: f, flag: 0}
}

func (wp *WorkerPool) Status() (bool, int, int) {
	var busy bool
	busyCount := atomic.LoadInt64(wp.busyCount)
	totalTasks := atomic.LoadInt64(wp.totalTasks)
	if busyCount == 0 && len(wp.remainingTasks) == 0 {
		busy = false
	} else {
		busy = true
	}
	return busy, len(wp.remainingTasks) + int(busyCount), int(totalTasks)
}

func (wp *WorkerPool) Close() {
	for *wp.currentWorkers > 0 {
		wp.remainingTasks <- task{flag: 1}
	}
}

func (wp *WorkerPool) Wait() {
	for *wp.busyCount > 0 {
		time.Sleep(time.Second)
	}
}