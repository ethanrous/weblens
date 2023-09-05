package util

import (
	"sync/atomic"
)

type workerPool struct {
	maxWorker int
	queuedTasks chan func()
	busyCount int64
}

func NewWorkerPool(workerCount int) (workerPool) {
	newWp := workerPool{workerCount, make(chan func(), workerCount * 2), int64(0)}
	return newWp
}

func (wp *workerPool) Run() {
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

func (wp *workerPool) AddTask(f func()) {
	wp.queuedTasks <- f
}

func (wp *workerPool) IsBusy() (bool, int, int) {
	var busy bool
	if wp.busyCount == 0 && len(wp.queuedTasks) == 0 {
		busy = false
	} else {
		busy = true
	}
	return busy, len(wp.queuedTasks), int(wp.busyCount)
}