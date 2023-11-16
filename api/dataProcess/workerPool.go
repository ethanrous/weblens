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
			defer func(){
				err := recover()
				if err != nil {
					util.Error.Println("Recovered panic at the surface of worker thread, this should have been handled sooner: ", err)
				}
			}()
			atomic.StoreInt64(wp.currentWorkers, atomic.AddInt64(wp.currentWorkers, 1))
			for task := range wp.remainingTasks {
				if task.flag == 1 {
					atomic.StoreInt64(wp.currentWorkers, atomic.AddInt64(wp.currentWorkers, -1))
					util.Debug.Println("Worker ", workerID, " Exiting")
					break
				}
				atomic.StoreInt64(wp.busyCount, atomic.AddInt64(wp.busyCount, 1))
				util.Debug.Printf("Worker %d starting work", workerID)
				task.work()
				util.Debug.Printf("Worker %d finished work", workerID)
				atomic.StoreInt64(wp.busyCount, atomic.AddInt64(wp.busyCount, -1))
			}
		}(i + 1)
	}
}

func (wp *WorkerPool) AddTask(f func()) {
	atomic.StoreInt64(wp.totalTasks, atomic.AddInt64(wp.totalTasks, 1))
	wp.remainingTasks <- task{work: f, flag: 0}
}

// Returns the count of tasks in the queue, and the total number of tasks accepted, number of busy workers, the total number of live workers in the worker pool
func (wp *WorkerPool) Status() (int, int, int, int) {

	busyCount := *wp.busyCount
	totalTasks := *wp.totalTasks
	currentWorkers := *wp.currentWorkers

	return len(wp.remainingTasks), int(totalTasks), int(busyCount), int(currentWorkers)
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