package task

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/modules/context"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var ErrTaskError = errors.New("task error")
var ErrTaskExit = errors.New("task exit")
var ErrTaskTimeout = errors.New("task timeout")

type hit struct {
	time   time.Time
	target *Task
}

type workChannel chan *Task
type hitChannel chan hit

type job struct {
	handler task_mod.CleanupFunc
	opts    TaskOptions
}

var _ task_mod.WorkerPool = (*WorkerPool)(nil)

type WorkerPool struct {
	ctx context.ContextZ

	busyCount *atomic.Int64 // Number of workers currently executing a task

	registeredJobs map[string]job

	taskMap map[string]*Task

	poolMap map[string]*TaskPool

	taskStream workChannel
	hitStream  hitChannel

	exitSignal chan bool

	retryBuffer    []*Task
	maxWorkers     atomic.Int64 // Max allowed worker count
	currentWorkers atomic.Int64 // Current total of workers on the pool

	lifetimeQueuedCount atomic.Int64

	exitFlag atomic.Int64

	jobsMu sync.RWMutex

	taskMu sync.RWMutex

	poolMu       sync.Mutex
	taskBufferMu sync.Mutex
}

func NewWorkerPool(ctx context.ContextZ, initWorkers int) *WorkerPool {
	if initWorkers == 0 {
		initWorkers = 1
	}

	newWp := &WorkerPool{
		ctx:            ctx,
		registeredJobs: map[string]job{},
		taskMap:        map[string]*Task{},
		poolMap:        map[string]*TaskPool{},

		busyCount: &atomic.Int64{},

		taskStream:  make(workChannel, initWorkers*1000),
		retryBuffer: []*Task{},

		hitStream: make(hitChannel, initWorkers*2),

		exitSignal: make(chan bool),
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
func (wp *WorkerPool) NewTaskPool(replace bool, createdBy task_mod.Task) task_mod.Pool {
	tp := wp.newTaskPoolInternal()
	if createdBy != nil {
		t := createdBy.(*Task)
		tp.createdBy = t
		if !createdBy.GetTaskPool().IsGlobal() {
			tp.createdBy = t
		}
	}
	if replace {
		wp.addReplacementWorker()
		tp.hasQueueThread = true
	}

	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	wp.poolMap[tp.ID()] = tp

	if createdBy != nil {
		createdBy.SetChildTaskPool(tp)
	}

	return tp
}

func (wp *WorkerPool) GetTaskPool(tpId string) *TaskPool {
	wp.poolMu.Lock()
	defer wp.poolMu.Unlock()
	return wp.poolMap[tpId]
}

func (wp *WorkerPool) GetTasksByJobName(jobName string) []*Task {
	wp.taskMu.RLock()
	defer wp.taskMu.RUnlock()

	var ret []*Task
	for _, t := range wp.taskMap {
		if t.JobName() == jobName {
			return append(ret, t)
		}
	}
	return ret
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

// RegisterJob adds a template for a repeatable job that can be called upon later in the program
func (wp *WorkerPool) RegisterJob(jobName string, fn task_mod.CleanupFunc, opts ...TaskOptions) {
	wp.jobsMu.Lock()
	defer wp.jobsMu.Unlock()

	o := TaskOptions{}
	if len(opts) != 0 {
		o = opts[0]
	}
	wp.registeredJobs[jobName] = job{handler: fn, opts: o}
}

func (wp *WorkerPool) DispatchJob(ctx context.ContextZ, jobName string, meta task_mod.TaskMetadata, pool *TaskPool) (*Task, error) {
	if meta.JobName() != jobName {
		return nil, errors.Errorf("job name does not match task metadata")
	}

	if err := meta.Verify(); err != nil {
		return nil, err
	}

	wp.jobsMu.RLock()
	if wp.registeredJobs[jobName].handler == nil {
		wp.jobsMu.RUnlock()
		return nil, errors.Errorf("trying to dispatch non-registered job: %s", jobName)
	}
	wp.jobsMu.RUnlock()

	if pool == nil {
		pool = wp.GetTaskPool("GLOBAL")
	}

	job := wp.getRegisteredJob(jobName)

	var taskId string
	if meta == nil {
		taskId = globbyHash(8, time.Now().String())
	} else {
		metaStr := meta.MetaString()
		if job.opts.Unique {
			metaStr += time.Now().String()
		}
		taskId = globbyHash(8, metaStr)
	}

	t := wp.GetTask(taskId)
	ctx.Log().UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("task_id", taskId).Str("job_name", jobName)
	})

	appCtx, ok := ctx.(context.AppContexter)
	if !ok {
		return nil, errors.New("context is not an AppContexter")
	} else {
		ctx = appCtx.AppCtx()
	}

	if t != nil {
		ctx.Log().Trace().Msgf("Task [%s] already exists, not launching again", taskId)
		return t, nil
	} else {
		t = &Task{
			taskId:   taskId,
			jobName:  jobName,
			metadata: meta,
			work:     job,

			queueState: PreQueued,

			// signal chan must be buffered so caller doesn't block trying to close many tasks
			signalChan: make(chan int, 1),

			waitChan: make(chan struct{}),

			Ctx: ctx,
		}
	}

	wp.addTask(t)

	ctx.Log().Trace().Stack().Err(errors.New("Task created")).Msgf("Task [%s] created", taskId)

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
			return nil, errors.Errorf("Attempted to re-queue a [%s] task that is already in another queue", t.jobName)
		}
		pool.addTask(t)
		return t, nil
	}

	if pool.allQueuedFlag.Load() {
		// We cannot add tasks to a queue that has been closed
		return nil, errors.Errorf("attempting to add task [%s] to closed task queue [pool created by %s]", t.JobName(), pool.ID())
	}

	pool.totalTasks.Add(1)

	if !pool.IsRoot() {
		pool.GetRootPool().IncTaskCount(1)
	}

	// Set the tasks queue
	t.setTaskPoolInternal(pool)

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

func (wp *WorkerPool) getRegisteredJob(jobName string) job {
	wp.jobsMu.RLock()
	defer wp.jobsMu.RUnlock()
	return wp.registeredJobs[jobName]
}

func (wp *WorkerPool) workerRecover(task *Task, workerId int64) {
	recovered := recover()
	if recovered != nil {
		// Make sure what we got is an error
		var err error
		switch e := recovered.(type) {
		case error:
			if errors.Is(err, ErrTaskError) {
				// wp.ctx.Log().Error.Printf("Task [%s] exited with an error", task.TaskId())
				// wp.ctx.Log().Error().Stack().Err(task.err).Msg("")
				return
			} else if errors.Is(err, ErrTaskExit) {
				return
			}
			err = errors.WithStack(e)
		default:
			err = errors.Errorf("%s", recovered)
		}
		// wp.ctx.Log().Raw.Printf(
		// 	"\n\tWorker %d recovered panic: \u001b[31m%s\u001B[0m\n\n%s\n", workerId, recovered,
		// 	werror.GetStack(2).String(),
		// )
		task.error(err)
	}
}

// saftyWork wraps the task execution with a recover, so if there are any panics
// during the task, we can catch them, display them, and safely remove the task.
func (wp *WorkerPool) safetyWork(task *Task, workerId int64) {
	defer wp.workerRecover(task, workerId)

	if task.exitStatus != task_mod.TaskNoStatus {
		task.Ctx.Log().Trace().Msgf("Task [%s] already has exit status [%s], not running", task.taskId, task.exitStatus)
	} else {
		task.work.handler(task)
	}

}

func (wp *WorkerPool) AddHit(time time.Time, target task_mod.Task) {
	wp.hitStream <- hit{time: time, target: target.(*Task)}
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
				log.Debug().Msg("Task reaper exiting")
				return
			}
			log.Warn().Msg("Reaper not exiting?")
		case newHit := <-wp.hitStream:
			go func(h hit) { time.Sleep(time.Until(h.time)); timerStream <- h.target }(newHit)
		case task := <-timerStream:
			// Possible that the task has kicked its timeout down the road
			// since it first queued it, we must check that we are past the
			// timeout before cancelling the task. Also check that it
			// has not already finished
			timeout := task.GetTimeout()
			if task.QueueState() != Exited && time.Until(timeout) <= 0 && timeout.Unix() != 0 {
				log.Warn().Msgf("Sending timeout signal to T[%s]\n", task.taskId)
				task.Cancel()
				task.error(ErrTaskTimeout)
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
	for wp.currentWorkers.Load() > 0 {
		log.Debug().Msgf("Waiting for %d workers to exit", wp.currentWorkers.Load())
		time.Sleep(100 * time.Millisecond)
	}
}

func (wp *WorkerPool) GetTask(taskId string) *Task {
	wp.taskMu.Lock()
	t := wp.taskMap[taskId]
	wp.taskMu.Unlock()
	return t
}

func (wp *WorkerPool) addTask(task *Task) {
	wp.taskMu.Lock()
	defer wp.taskMu.Unlock()
	wp.taskMap[task.Id()] = task
}

func (wp *WorkerPool) removeTask(taskId string) {
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
		wp.ctx.Log().Trace().Msgf("Spinning up worker with id [%d] o7", workerId)
		defer func() {
			wp.ctx.Log().Debug().Msgf("worker %d exiting, %d workers remain", workerId, wp.currentWorkers.Add(-1))
		}()

		// WorkLoop:
		for {
			select {
			case _, ok := <-wp.exitSignal:
				if !ok {
					return
				}
			case t := <-wp.taskStream:
				{
					// Even if we get a new task, but the wp is marked for exit, we just exit
					if wp.exitFlag.Load() == 1 {
						// We don't care about the exitLock here, since the whole wp
						// is going down anyway.

						return
					}

					// Replacement workers are not allowed to do "scan_directory" tasks
					// TODO - generalize
					if replacement && t.jobName == "scan_directory" && t.exitStatus == task_mod.TaskNoStatus {

						// If there are twice the number of free spaces in the chan, don't bother pulling
						// everything into the waiting buffer, just put it at the end right now.
						if cap(wp.taskStream)-len(wp.taskStream) > int(wp.currentWorkers.Load())*2 {
							// util.Debug().Msg("Replacement worker putting scan dir task back")
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

					t.Ctx.Log().UpdateContext(func(c zerolog.Context) zerolog.Context {
						return c.Int("worker_id", int(workerId))
					})

					// Inc tasks being processed
					wp.busyCount.Add(1)
					t.Ctx.Log().Trace().Func(func(e *zerolog.Event) { e.Msgf("Starting task") })
					wp.safetyWork(t, workerId)
					t.Ctx.Log().Trace().Func(func(e *zerolog.Event) {
						t.updateMu.RLock()
						defer t.updateMu.RUnlock()
						e.Dur("task_duration", t.ExeTime()).Str("exit_status", string(t.exitStatus)).Msg("Task finished")
					})

					// Dec tasks being processed
					wp.busyCount.Add(-1)

					// Tasks must set their completed flag before exiting
					// if it wasn't done in the work body, we do it for them
					t.updateMu.Lock()
					if t.queueState != Exited {
						t.Success("closed by worker pool")
					}
					t.updateMu.Unlock()

					result := t.GetResults()
					t.updateMu.Lock()
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
					t.updateMu.Unlock()
					close(t.waitChan)

					// Potentially find the task pool that houses this task pool. All child
					// task pools report their status to the root task pool as well.
					// Do not use any global pool as the root
					rootTaskPool := t.taskPool.GetRootPool()

					if t.exitStatus == task_mod.TaskError && !rootTaskPool.IsGlobal() {
						rootTaskPool.AddError(t)
					}

					if !t.work.opts.Persistent {
						wp.removeTask(t.taskId)
					}

					var canContinue bool
					directParent := t.GetTaskPool()

					directParent.IncCompletedTasks(1)

					// Run the cleanup routine for errors, if any
					if t.exitStatus == task_mod.TaskError || t.exitStatus == task_mod.TaskCanceled {
						t.Ctx.Log().Trace().Msgf("Running error cleanup")
						for _, ecf := range t.errorCleanup {
							ecf(t)
						}
						t.Ctx.Log().Trace().Msgf("Finished error cleanup")
					}

					// We want the pool completed and error count to reflect the task has completed
					// when we are doing cleanup. Cleanup is intended to execute "after" the task
					// has finished, so we must inc the completed tasks counter (above) before cleanup
					t.Ctx.Log().Trace().Msgf("Running cleanup(s)")
					t.updateMu.RLock()
					cleanups := t.cleanups
					t.updateMu.RUnlock()
					for _, cf := range cleanups {
						cf(t)
					}
					t.Ctx.Log().Trace().Msgf("Finished cleanup")

					// The post action is intended to run after the task has completed, and after
					// the cleanup has run. This is useful for tasks that need to use the result of
					// the task to do something else, but don't want to block until the task is done
					if t.postAction != nil && t.exitStatus == task_mod.TaskSuccess {
						t.Ctx.Log().Trace().Msg("Running task post-action")
						t.postAction(t.result)
						t.Ctx.Log().Trace().Msg("Finished task post-action")
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
						canContinue = directParent.HandleTaskExit(replacement)

						directParent.UnlockExit()
					} else {
						rootTaskPool.IncCompletedTasks(1)

						// Must hold both locks (and must acquire root lock first) to enter a dual-update.
						// Any other ordering will result in race conditions or deadlocks
						rootTaskPool.LockExit()

						directParent.LockExit()
						uncompletedTasks := directParent.GetTotalTaskCount() - directParent.GetCompletedTaskCount()
						wp.ctx.Log().Debug().Msgf(
							"Uncompleted tasks on tp created by %s: %d",
							directParent.CreatedInTask().Id(), uncompletedTasks-1,
						)
						canContinue = directParent.HandleTaskExit(replacement)

						// We *should* get the same canContinue value from here, so we do not
						// check it a second time. If we *don't* get the same value, we can safely ignore it
						rootTaskPool.HandleTaskExit(replacement)

						directParent.UnlockExit()
						rootTaskPool.UnlockExit()
					}

					if !canContinue {
						return
					}
				}
			}
		}
	}(wp.currentWorkers.Add(1) - 1)
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
		wp.ctx.Log().Error().Err(err).Msg("Failed to generate UUID for new task pool")
		return nil
	}

	newQueue := &TaskPool{
		id:           tpId.String(),
		tasks:        map[string]*Task{},
		workerPool:   wp,
		createdAt:    time.Now(),
		erroredTasks: make([]task_mod.Task, 0),
		waiterGate:   make(chan struct{}),
		log:          wp.ctx.Log().With().Str("task_pool", tpId.String()).Logger(),
	}

	return newQueue
}

func (wp *WorkerPool) removeTaskPool(tpId string) {
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
	for wp.exitFlag.Load() == 0 {
		time.Sleep(time.Second * waitTime)
		remaining, total, busy, alive, retrySize := wp.Status()
		if lastCount != remaining {
			lastCount = remaining
			waitTime = 1
			wp.ctx.Log().Debug().Int("queue_remaining", remaining).Int("queue_total", total).Int("queue_buffered", retrySize).Int("busy_workers", busy).Int("alive_workers", alive).Msg(
				"Worker pool status",
			)
		} else if waitTime < time.Second*10 {
			waitTime += 1
		}
	}
	wp.ctx.Log().Warn().Msg("status reporter exited")
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

	wp.ctx.Log().Debug().Msg("buffer drainer exited")
}

func (wp *WorkerPool) addToRetryBuffer(tasks ...*Task) {
	wp.taskBufferMu.Lock()
	defer wp.taskBufferMu.Unlock()
	wp.retryBuffer = append(wp.retryBuffer, tasks...)
}
