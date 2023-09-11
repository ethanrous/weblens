package util

import (
	"sync/atomic"
)

type WorkerPool struct {
	maxWorker int
	remainingTasks chan func()
	busyCount *int64
	totalTasks *int64
}

func NewWorkerPool(workerCount int) (WorkerPool) {
	busyCount := int64(0)
	totalTasks := int64(0)
	newWp := WorkerPool{workerCount, make(chan func(), workerCount * 1000), &busyCount, &totalTasks}
	return newWp
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		go func(workerID int) {
			for task := range wp.remainingTasks {
				atomic.StoreInt64(wp.busyCount, atomic.AddInt64(wp.busyCount, 1))
				task()
				atomic.StoreInt64(wp.busyCount, atomic.AddInt64(wp.busyCount, -1))
			}
		}(i + 1)
	}
}

func (wp *WorkerPool) AddTask(f func()) {
	atomic.StoreInt64(wp.totalTasks, atomic.AddInt64(wp.totalTasks, 1))
	wp.remainingTasks <- f
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