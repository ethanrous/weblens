package util

import (
	"sync/atomic"
)

type WorkerPool struct {
	maxWorker int
	queuedTasks chan func()
	busyCount int64
}

func NewWorkerPool(workerCount int) (WorkerPool) {
	newWp := WorkerPool{workerCount, make(chan func(), workerCount * 2), int64(0)}
	return newWp
}

func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		go func(workerID int) {
			for task := range wp.queuedTasks {
				atomic.AddInt64(&wp.busyCount, 1)
				task()
				atomic.AddInt64(&wp.busyCount, -1)
			}
		}(i + 1)
	}
}

func (wp *WorkerPool) AddTask(f func()) {
	wp.queuedTasks <- f
}

func (wp *WorkerPool) IsBusy() (bool, int, int) {
	var busy bool
	if wp.busyCount == 0 && len(wp.queuedTasks) == 0 {
		busy = false
	} else {
		busy = true
	}
	return busy, len(wp.queuedTasks), int(wp.busyCount)
}