package task

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/metrics"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/google/uuid"
)

type TaskHandler func(*Task)

type hit struct {
	time   time.Time
	target *Task
}

type workChannel chan *Task
type hitChannel chan hit

type WorkerPool struct {
	maxWorkers     atomic.Int64  // Max allowed worker count
	currentWorkers atomic.Int64  // Current total of workers on the pool
	busyCount      *atomic.Int64 // Number of workers currently executing a task

	lifetimeQueuedCount atomic.Int64

	jobsMu         sync.RWMutex
	registeredJobs map[string]TaskHandler

	taskMu  sync.Mutex
	taskMap map[TaskId]*Task

	poolMu  sync.Mutex
	poolMap map[TaskId]*TaskPool

	taskStream   workChannel
	taskBufferMu sync.Mutex
	retryBuffer []*Task
	hitStream    hitChannel

	exitFlag   atomic.Int64
	exitSignal chan bool

	logLevel int
}

func NewWorkerPool(initWorkers int, logLevel int) *WorkerPool {
	if initWorkers == 0 {
		initWorkers = 1
	}
	log.Debug.Printf("Starting new worker pool with %d workers", initWorkers)

	newWp := &WorkerPool{
		registeredJobs: map[string]TaskHandler{},
		taskMap:        map[TaskId]*Task{},
		poolMap:        map[TaskId]*TaskPool{},

		busyCount: &atomic.Int64{},

		taskStream:  make(workChannel, initWorkers*1000),
		retryBuffer: []*Task{},

		hitStream: make(hitChannel, initWorkers*2),

		exitSignal: make(chan bool),

		logLevel: logLevel,
	}
	// Worker pool starts disabled
	newWp.exitFlag.Store(1)

	newWp.maxWorkers.Store(int64(initWorkers))

	// Worker pool always has one global queue
	globalPool := newWp.newTaskPoolInternal()
	globalPool.id = "GLOBAL"
	globalPool.MarkGlobal()

	newWp.poolMap["GLOBAL"] = globalPool

	return newWp
}

// NewTaskPool `replace` spawns a temporary replacement thread on the parent worker pool.
// This prevents a deadlock when the queue fills up while adding many tasks, and none are being executed
//
// `parent` allows chaining of task pools for floating updates to the top. This makes
// it possible for clients to subscribe to a single task, and get notified about
// all of the sub-updates of that task
// See taskPool.go
func (wp *WorkerPool) NewTaskPool(replace bool, createdBy *Task) *TaskPool {
	tp := wp.newTaskPoolInternal()
	if createdBy != nil {
		tp.createdBy = createdBy
		if !createdBy.GetTaskPool().IsGlobal() {
			tp.createdBy = createdBy
		}
	}
	if replace {
		wp.addReplacementWorker()
		tp.hasQueueThread = true
	}

	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	wp.poolMap[tp.ID()] = tp

	return tp
}

func (wp *WorkerPool) GetTaskPool(tpId TaskId) *TaskPool {
	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	return wp.poolMap[tpId]
}

func (wp *WorkerPool) GetTaskPoolByJobName(jobName string) *TaskPool {
	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	for _, tp := range wp.poolMap {
		if tp.CreatedInTask() != nil && tp.CreatedInTask().JobName() == jobName {
			return tp
		}
	}
	return nil
}

/* RegisterJob adds a template for a repeatable job that can be called upon later in the program */
func (wp *WorkerPool) RegisterJob(jobName string, fn TaskHandler) {
	wp.taskMu.Lock()
	defer wp.taskMu.Unlock()
	wp.registeredJobs[jobName] = fn
}

func (wp *WorkerPool) DispatchJob(jobName string, meta TaskMetadata, pool *TaskPool) (*Task, error) {
	if meta.JobName() != jobName {
		return nil, errors.New("job name does not match task metadata")
	}

	if err := meta.Verify(); err != nil {
		return nil, err
	}

	if wp.registeredJobs[jobName] == nil {
		return nil, werror.Errorf("trying to dispatch non-registered job: %s", jobName)
	}

	if pool == nil {
		pool = wp.GetTaskPool("GLOBAL")
	}

	var taskId TaskId
	if meta == nil {
		taskId = TaskId(globbyHash(8, time.Now().String()))
	} else {
		taskId = TaskId(globbyHash(8, meta.MetaString()))
	}

	t := wp.GetTask(taskId)
	if t != nil {
		// if jobName == "write_file" {
		// 	existingTask.ClearAndRecompute()
		// }
		return t, nil
	} else {
		t = &Task{
			taskId:   taskId,
			jobName:  jobName,
			metadata: meta,
			work:     wp.registeredJobs[jobName],

			queueState: PreQueued,

			// signal chan must be buffered so caller doesn't block trying to close many tasks
			signalChan: make(chan int, 1),

			sw: internal.NewStopwatch(fmt.Sprintf("%s Task [%s]", jobName, taskId)),
		}
	}

	// Lock the waiter gate immediately. The task cleanup routine will clear
	// this lock when the task exits, which will allow any thread waiting on
	// the task to return
	t.waitMu.Lock()

	wp.addTask(t)

	if wp.exitFlag.Load() == 1 {
		return nil, errors.New("not queuing task while worker pool is going down")
	}

	if t.err != nil {
		// Tasks that have failed will not be re-tried. If the errored task is removed from the
		// task map, then it will be re-tried because the previous error was lost. This can be
		// sometimes be useful, some tasks auto-remove themselves after they finish.

		return nil, errors.New("Not re-queueing task that has error set")
	}

	if t.taskPool != nil && (t.taskPool != pool || t.queueState != PreQueued) {
		// Task is already queued, we are not allowed to move it to another queue.
		// We can call .ClearAndRecompute() on the task and it will queue it
		// again, but it cannot be transferred
		if t.taskPool != pool {
			return nil, werror.Errorf("Attempted to re-queue a [%s] task that is already in another queue", t.jobName)
		}
		pool.addTask(t)
		return t, nil
	}

	if pool.allQueuedFlag.Load() {
		// We cannot add tasks to a queue that has been closed
		return nil, werror.WithStack(errors.New("attempting to add task to closed task queue"))
	}

	pool.totalTasks.Add(1)

	if !pool.IsRoot() {
		pool.GetRootPool().totalTasks.Add(1)
	}

	// Set the tasks queue
	t.taskPool = pool

	wp.lifetimeQueuedCount.Add(1)

	// Put the task in the queue
	t.queueState = InQueue
	if len(wp.retryBuffer) != 0 || len(wp.taskStream) == cap(wp.taskStream) {
		wp.addToRetryBuffer(t)
	} else {
		wp.taskStream <- t
	}

	pool.addTask(t)

	return t, nil
}

func (wp *WorkerPool) workerRecover(task *Task, workerId int64) {
	recovered := recover()
	if recovered != nil {
		// Make sure what we got is an error
		switch err := recovered.(type) {
		case error:
			if err.Error() == werror.ErrTaskError.Error() {
				if wp.logLevel != -1 {
					log.Error.Printf("Task [%s] exited with an error", task.TaskId())
					log.ErrTrace(task.err)
				}
				return
			} else if err.Error() == werror.ErrTaskExit.Error() {
				return
			}
		default:
			recovered = errors.New(fmt.Sprint(recovered))
		}
		if wp.logLevel != -1 {
			log.ErrorCatcher.Printf("Worker %d recovered error: %s\n%s\n", workerId, recovered, debug.Stack())
		}
		task.error(recovered.(error))
	}
}

// saftyWork wraps the task execution with a recover, so if there are any panics
// during the task, we can catch them, display them, and safely remove the task.
func (wp *WorkerPool) safetyWork(task *Task, workerId int64) {
	defer wp.workerRecover(task, workerId)
	task.work(task)

	task.updateMu.Lock()
	if task.postAction != nil && task.exitStatus == TaskSuccess {
		task.updateMu.Unlock()
		task.postAction(task.result)
		return
	}
	task.updateMu.Unlock()
}

func (wp *WorkerPool) AddHit(time time.Time, target *Task) {
	wp.hitStream <- hit{time: time, target: target}
}

// The reaper handles timed cancellation of tasks. If a task might
// wait on some action from the client, it should queue a timeout with
// the reaper so it doesn't hang forever
func (wp *WorkerPool) reaper() {
	timerStream := make(chan *Task)
	for wp.exitFlag.Load() == 0 {
		select {
		case _, ok := <-wp.exitSignal:
			if !ok {
				log.Debug.Println("reaper exiting")
				return
			}
			log.Debug.Println("reaper not exiting?")
		case newHit := <-wp.hitStream:
			go func(h hit) { time.Sleep(time.Until(h.time)); timerStream <- h.target }(newHit)
		case task := <-timerStream:
			// Possible that the task has kicked it's timeout down the road
			// since it first queued it, we must check that we are past the
			// timeout before cancelling the task. Also check that it
			// has not already finished
			if task.queueState != Exited && time.Until(task.timeout) <= 0 && task.timeout.Unix() != 0 {
				log.Warning.Printf("Sending timeout signal to T[%s]\n", task.taskId)
				task.Cancel()
				task.error(werror.ErrTaskTimeout)
			}
		}
	}
}

// Run launches the standard threads for this worker pool
func (wp *WorkerPool) Run() {
	wp.exitFlag.Store(0)

	// Spawn the timeout checker
	go wp.reaper()

	// Spawn the buffer worker
	go wp.bufferDrainer()

	// Spawn the status printer
	go wp.statusReporter()

	var i int64
	for i = 0; i < wp.maxWorkers.Load(); i++ {
		// These are the base, 'omnipresent' threads for this pool,
		// so they are NOT replacement workers. See wp.execWorker
		// for more info
		wp.execWorker(false)
	}
}

func (wp *WorkerPool) Stop() {
	wp.exitFlag.Store(1)
	close(wp.exitSignal)
	for wp.currentWorkers.Load() != 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func (wp *WorkerPool) GetTask(taskId TaskId) *Task {
	wp.taskMu.Lock()
	t := wp.taskMap[taskId]
	wp.taskMu.Unlock()
	return t
}

func (wp *WorkerPool) addTask(task *Task) {
	wp.taskMu.Lock()
	defer wp.taskMu.Unlock()
	wp.taskMap[task.TaskId()] = task
}

func (wp *WorkerPool) removeTask(taskId TaskId) {
	wp.taskMu.Lock()
	t := wp.taskMap[taskId]
	wp.taskMu.Unlock()
	if t == nil {
		return
	}

	t.GetTaskPool().RemoveTask(taskId)

	wp.taskMu.Lock()
	defer wp.taskMu.Unlock()
	delete(wp.taskMap, taskId)
}

// Main worker method, spawn a worker and loop over the task channel
//
// `replacement` specifies if the worker is a temporary replacement for another
// task that is parking for a long time. Replacement workers behave a bit
// different to minimize parked time of the other task.
func (wp *WorkerPool) execWorker(replacement bool) {
	go func(workerId int64) {
		wp.currentWorkers.Add(1)

		// Inc alive workers
		defer wp.currentWorkers.Add(-1)
		// util.Debug.Printf("Worker %d reporting for duty o7", workerId)

		// WorkLoop:
		for {
			select {
			case _, ok := <-wp.exitSignal:
				if !ok {
					log.Debug.Printf("worker %d exiting", workerId)
					return
				}
			case t := <-wp.taskStream:
				{
					// Even if we get a new task, if the wp is marked for exit, we just exit
					if wp.exitFlag.Load() == 1 {
						// We don't care about the exitLock here, since the whole wp
						// is going down anyway.

						return
					}

					// Replacement workers are not allowed to do "scan_directory" tasks
					// TODO - generalize
					if replacement && t.jobName == "scan_directory" {

						// If there are twice the number of free spaces in the chan, don't bother pulling
						// everything into the waiting buffer, just put it at the end right now.
						if cap(wp.taskStream)-len(wp.taskStream) > int(wp.currentWorkers.Load())*2 {
							// util.Debug.Println("Replacement worker putting scan dir task back")
							wp.taskStream <- t
							continue
						}
						tBuf := []*Task{t}
						for t = range wp.taskStream {
							if t.jobName == "scan_directory" {
								tBuf = append(tBuf, t)
							} else {
								break
							}
						}
						wp.addToRetryBuffer(tBuf...)
					}

					// Inc tasks being processed
					wp.busyCount.Add(1)
					metrics.BusyWorkerGuage.Inc()
					t.SwLap("Task start")
					log.Trace.Printf("Starting %s task T[%s]", t.jobName, t.taskId)
					wp.safetyWork(t, workerId)
					t.SwLap("Task finish")
					t.sw.Stop()
					log.Trace.Printf("Finished %s task T[%s] in %s", t.jobName, t.taskId, t.ExeTime())

					// Dec tasks being processed
					wp.busyCount.Add(-1)
					metrics.BusyWorkerGuage.Dec()

					// Tasks must set their completed flag before exiting
					// if it wasn't done in the work body, we do it for them
					if t.queueState != Exited {
						t.Success("closed by worker pool")
					}

					result := t.GetResults()
					result["task_id"] = t.taskId
					result["exit_status"] = t.exitStatus
					wp.poolMu.Lock()
					var complete int64
					for _, p := range wp.poolMap {
						status := p.Status()
						complete += status.Complete
					}
					wp.poolMu.Unlock()
					result["queue_remaining"] = complete
					result["queue_total"] = wp.lifetimeQueuedCount.Load()

					// Wake any waiters on this task
					t.waitMu.Unlock()

					// Potentially find the task pool that houses this task pool. All child
					// task pools report their status to the root task pool as well.
					// Do not use any global pool as the root
					rootTaskPool := t.taskPool.GetRootPool()

					if t.exitStatus == TaskError && !rootTaskPool.IsGlobal() {
						rootTaskPool.AddError(t)
					}

					if !t.persistent {
						wp.removeTask(t.taskId)
					}

					canContinue := true
					directParent := t.GetTaskPool()

					directParent.completedTasks.Add(1)

					// We want the pool completed and error count to reflec the task has completed
					// when we are doing cleanup. Cleanup is intended to execute "after" the task
					// has finished, so we must inc the completed tasks counter (above) before cleanup
					if t.cleanup != nil {
						t.cleanup(t)
						t.cleanup = nil
					}

					if directParent.IsRoot() {
						// Updating the number of workers and then checking it's value is dangerous
						// to do concurrently. Specifically, the waiterGate lock on the queue will,
						// very rarely, attempt to unlock twice if two tasks finish at the same time.
						// So we must treat this whole area as a critical section
						directParent.LockExit()

						// Set values and notifications now that task has completed. Returns
						// a bool that specifies if this thread should continue and grab another
						// task, or if it should exit
						canContinue = directParent.handleTaskExit(replacement)

						directParent.UnlockExit()
					} else {
						rootTaskPool.completedTasks.Add(1)

						// Must hold both locks (and must acquire root lock first) to enter a dual-update.
						// Any other ordering will result in race conditions or deadlocks
						rootTaskPool.LockExit()

						directParent.LockExit()
						uncompletedTasks := directParent.totalTasks.Load() - directParent.completedTasks.Load()
						log.Debug.Printf(
							"Uncompleted tasks on tp created by %s: %d",
							directParent.CreatedInTask().TaskId(), uncompletedTasks-1,
						)
						canContinue = directParent.handleTaskExit(replacement)

						// We *should* get the same canContinue value from here, so we do not
						// check it a second time. If we *don't* get the same value, we can safely ignore it
						rootTaskPool.handleTaskExit(replacement)

						directParent.UnlockExit()
						rootTaskPool.UnlockExit()
					}

					if !canContinue {
						return
					}
				}
			}
		}
	}(wp.currentWorkers.Load())
}

/*
We can create "replacement" workers as to not create a deadlock.
e.g. all n `wp.maxWorkers` are "scan_directory" tasks parked and waiting
for their "scan_file" children tasks to complete, but never will since
all worker threads are taken up by blocked tasks.

Replacement workers:
1. Are not allowed to pick up "scan_directory" tasks
2. will exit once the main thread has woken up, and shrunk the worker pool back down
*/
func (wp *WorkerPool) addReplacementWorker() {
	wp.maxWorkers.Add(1)
	wp.execWorker(true)
}

func (wp *WorkerPool) removeWorker() {
	wp.maxWorkers.Add(-1)
}

func (wp *WorkerPool) newTaskPoolInternal() *TaskPool {
	tpId, err := uuid.NewUUID()
	if err != nil {
		log.ShowErr(err)
		return nil
	}

	newQueue := &TaskPool{
		id:           TaskId(tpId.String()),
		tasks:        map[TaskId]*Task{},
		workerPool:   wp,
		createdAt:    time.Now(),
		erroredTasks: make([]*Task, 0),
	}

	// The waiterGate begins locked, and will only unlock when all
	// tasks have been queued and finish, then the waiters are released
	newQueue.waiterGate.Lock()

	return newQueue
}

func (wp *WorkerPool) removeTaskPool(tpId TaskId) {
	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	delete(wp.poolMap, tpId)
}

// Status returns the count of tasks in the queue, the total number of tasks accepted,
// number of busy workers, and the total number of live workers in the worker pool
func (wp *WorkerPool) Status() (int, int, int, int, int) {
	total := int(wp.lifetimeQueuedCount.Load())
	wp.taskBufferMu.Lock()
	defer wp.taskBufferMu.Unlock()

	return len(wp.taskStream), total, int(wp.busyCount.Load()), int(wp.currentWorkers.Load()), len(wp.retryBuffer)
}

func (wp *WorkerPool) statusReporter() {
	var lastCount int
	var waitTime time.Duration = 1
	for {
		time.Sleep(time.Second * waitTime)
		remaining, total, busy, alive, retrySize := wp.Status()
		if lastCount != remaining {
			lastCount = remaining
			waitTime = 1
			log.Info.Printf(
				"Task worker pool status : Queued[%d]/Total[%d], Buffered[%d], Busy[%d], Alive[%d]",
				remaining, total, retrySize, busy, alive,
			)
		} else if waitTime < time.Second*10 {
			waitTime += 1
		}
	}
}

func (wp *WorkerPool) bufferDrainer() {
	for wp.exitFlag.Load() == 0 {
		wp.taskBufferMu.Lock()
		if len(wp.retryBuffer) != 0 && len(wp.taskStream) == 0 {
			for _, t := range wp.retryBuffer {
				wp.taskStream <- t
			}
			wp.retryBuffer = []*Task{}
		}
		wp.taskBufferMu.Unlock()
		time.Sleep(time.Second * 10)
	}

	log.Debug.Println("buffer drainer exited")
}

func (wp *WorkerPool) addToRetryBuffer(tasks ...*Task) {
	wp.taskBufferMu.Lock()
	defer wp.taskBufferMu.Unlock()
	wp.retryBuffer = append(wp.retryBuffer, tasks...)
}

type TaskService interface {
	RegisterJob(jobName string, fn TaskHandler)
	NewTaskPool(replace bool, createdBy *Task) *TaskPool

	GetTask(taskId TaskId) *Task
	GetTaskPool(TaskId) *TaskPool

	DispatchJob(jobName string, meta TaskMetadata, pool *TaskPool) (*Task, error)
}
