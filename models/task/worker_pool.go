package task

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var ErrTaskError = errors.New("task error")
var ErrTaskExit = errors.New("task exit")
var ErrTaskTimeout = errors.New("task timeout")
var ErrTaskCancelled = errors.New("task cancelled")
var ErrTaskAlreadyComplete = errors.New("task already complete")

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
	ctx context_mod.ContextZ

	busyCount *atomic.Int64 // Number of workers currently executing a task

	registeredJobs map[string]job

	taskMap map[string]*Task

	poolMap map[string]*TaskPool

	taskStream workChannel
	hitStream  hitChannel

	retryBuffer    []*Task
	maxWorkers     atomic.Int64 // Max allowed worker count
	currentWorkers atomic.Int64 // Current total of workers on the pool

	lifetimeQueuedCount atomic.Int64

	jobsMu sync.RWMutex

	taskMu sync.RWMutex

	poolMu       sync.Mutex
	taskBufferMu sync.Mutex
}

func NewWorkerPool(ctx context_mod.ContextZ, initWorkers int) *WorkerPool {
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
	}

	newWp.maxWorkers.Store(int64(initWorkers))

	// Worker pool always has one global queue
	globalPool := newWp.newTaskPoolInternal()
	globalPool.id = GlobalTaskPoolId
	globalPool.MarkGlobal()

	newWp.poolMap[GlobalTaskPoolId] = globalPool

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
		// We want to use the same context as the worker pool, but we need a cancelable context for when the task pool is closed
		ctx, cancel := context.WithCancel(wp.ctx)

		tp.AddCleanup(func(p task_mod.Pool) { cancel() })
		wp.addReplacementWorker(ctx)

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

func (wp *WorkerPool) DispatchJob(ctx context.Context, jobName string, meta task_mod.TaskMetadata, pool task_mod.Pool) (task_mod.Task, error) {
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
		pool = wp.GetTaskPool(GlobalTaskPoolId)
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

	if t != nil {
		t.Log().Debug().Msgf("Task [%s] already exists, not re-queueing again", taskId)
		return t, nil
	} else {
		newl := log.FromContext(ctx).With().
			Str("task_id", taskId).
			Str("job_name", jobName).
			Logger()
		ctx = log.WithContext(ctx, &newl)

		ctx, cancel := context.WithCancelCause(ctx)

		t = &Task{
			taskId:   taskId,
			jobName:  jobName,
			metadata: meta,
			work:     job,

			queueState: Created,

			// signal chan must be buffered so caller doesn't block trying to close many tasks
			waitChan: make(chan struct{}),

			Ctx:        ctx,
			cancelFunc: cancel,
		}
	}

	wp.addTask(t)

	t.Log().Trace().Stack().Msgf("Task [%s] created", taskId)

	select {
	case _, ok := <-t.Ctx.Done():
		if !ok {
			return nil, errors.New("not queuing task while worker pool is going down")
		}
	default:
	}

	if t.err != nil {
		// Tasks that have failed will not be re-tried. If the errored task is removed from the
		// task map, then it will be re-tried because the previous error was lost. This can be
		// sometimes be useful, some tasks auto-remove themselves after they finish.

		return nil, errors.New("Not re-queueing task that has error set")
	}

	tpool, ok := pool.(*TaskPool)
	if !ok {
		return nil, errors.Errorf("Task pool is not a TaskPool")
	}

	if t.taskPool != nil && (t.taskPool != tpool || t.queueState != Created) {
		// Task is already queued, we are not allowed to move it to another queue.
		// We can call .ClearAndRecompute() on the task and it will queue it
		// again, but it cannot be transferred
		if t.taskPool != tpool {
			return nil, errors.Errorf("Attempted to re-queue a [%s] task that is already in another queue", t.jobName)
		}
		tpool.addTask(t)
		return t, nil
	}

	if tpool.allQueuedFlag.Load() {
		// We cannot add tasks to a queue that has been closed
		return nil, errors.Errorf("attempting to add task [%s] to closed task queue [pool created by %s]", t.JobName(), tpool.ID())
	}

	tpool.totalTasks.Add(1)

	if !tpool.IsRoot() {
		tpool.GetRootPool().IncTaskCount(1)
	}

	// Set the tasks queue
	t.setTaskPoolInternal(tpool)

	wp.lifetimeQueuedCount.Add(1)

	// Put the task in the queue
	t.queueState = InQueue
	if len(wp.retryBuffer) != 0 || len(wp.taskStream) == cap(wp.taskStream) {
		wp.addToRetryBuffer(t)
	} else {
		wp.taskStream <- t
	}

	tpool.addTask(t)

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
			if errors.Is(e, ErrTaskError) {
				return
			} else if errors.Is(e, ErrTaskExit) {
				return
			}
			err = errors.WithStack(e)
		default:
			err = errors.Errorf("%s", recovered)
		}

		task.error(err)
	}
}

// saftyWork wraps the task execution with a recover, so if there are any panics
// during the task, we can catch them, display them, and safely remove the task.
func (wp *WorkerPool) safetyWork(task *Task, workerId int64) {
	defer wp.workerRecover(task, workerId)

	if task.exitStatus != task_mod.TaskNoStatus {
		log.FromContext(task.Ctx).Trace().Msgf("Task [%s] already has exit status [%s], not running", task.taskId, task.exitStatus)
	} else {
		task.StartTime = time.Now()
		task.work.handler(task)
		task.FinishTime = time.Now()
	}

}

func (wp *WorkerPool) cleanup(task *Task, workerId int64) {
	defer wp.workerRecover(task, workerId)

	// Run the cleanup routine for errors, if any
	if task.exitStatus == task_mod.TaskError || task.exitStatus == task_mod.TaskCanceled {
		log.FromContext(task.Ctx).Trace().Msgf("Running error cleanup")

		for _, ecf := range task.errorCleanup {
			ecf(task)
		}
		log.FromContext(task.Ctx).Trace().Msgf("Finished error cleanup")
	}

	// We want the pool completed and error count to reflect the task has completed
	// when we are doing cleanup. Cleanup is intended to execute "after" the task
	// has finished, so we must inc the completed tasks counter (above) before cleanup
	log.FromContext(task.Ctx).Trace().Msgf("Running cleanup(s)")
	task.updateMu.RLock()
	cleanups := task.cleanups
	task.updateMu.RUnlock()

	for _, cf := range cleanups {
		cf(task)
	}

	log.FromContext(task.Ctx).Trace().Msgf("Finished cleanup")

	// The post action is intended to run after the task has completed, and after
	// the cleanup has run. This is useful for tasks that need to use the result of
	// the task to do something else, but don't want to block until the task is done
	if task.postAction != nil && task.exitStatus == task_mod.TaskSuccess {
		log.FromContext(task.Ctx).Trace().Msg("Running task post-action")
		task.postAction(task.result)
		log.FromContext(task.Ctx).Trace().Msg("Finished task post-action")
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
	for {
		select {
		case _, ok := <-wp.ctx.Done():
			if !ok {
				log.GlobalLogger().Debug().Msg("Task reaper exiting")
				return
			}
			log.GlobalLogger().Warn().Msg("Reaper not exiting?")
		case newHit := <-wp.hitStream:
			go func(h hit) { time.Sleep(time.Until(h.time)); timerStream <- h.target }(newHit)
		case task := <-timerStream:
			// Possible that the task has kicked its timeout down the road
			// since it first queued it, we must check that we are past the
			// timeout before cancelling the task. Also check that it
			// has not already finished
			timeout := task.GetTimeout()
			if task.QueueState() != Exited && time.Until(timeout) <= 0 && timeout.Unix() != 0 {
				log.GlobalLogger().Warn().Msgf("Sending timeout signal to T[%s]\n", task.taskId)
				task.Cancel()
				task.error(ErrTaskTimeout)
			}
		}
	}
}

// Run launches the standard threads for this worker pool
func (wp *WorkerPool) Run() {
	// Spawn the timeout checker
	go wp.reaper()

	// Spawn the buffer worker
	go wp.bufferDrainer()

	// Spawn the status printer
	go wp.statusReporter()

	for range wp.maxWorkers.Load() {
		// These are the base, 'omnipresent' threads for this pool,
		// so they are NOT replacement workers. See wp.execWorker
		// for more info
		wp.execWorker(wp.ctx, false)
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
func (wp *WorkerPool) execWorker(ctx context.Context, replacement bool) {
	go func(workerId int64) {
		wp.ctx.Log().Debug().Msgf("Spinning up worker with id [%d] o7", workerId)

		defer func() {
			wp.ctx.Log().Debug().Msgf("worker %d exiting, %d workers remain", workerId, wp.currentWorkers.Add(-1))
		}()

		// WorkLoop:
		for {
			select {
			case _, ok := <-ctx.Done():
				if !ok {
					return
				}
			case t := <-wp.taskStream:
				{
					// Replacement workers are not allowed to do "scan_directory" tasks
					// TODO: - generalize
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

					newl := log.FromContext(t.Ctx).With().
						Int("worker_id", int(workerId)).
						Logger()

					t.Ctx = log.WithContext(t.Ctx, &newl)
					l := log.FromContext(t.Ctx)

					// Inc tasks being processed
					wp.busyCount.Add(1)
					l.Trace().Func(func(e *zerolog.Event) { e.Msgf("Starting task") })
					wp.safetyWork(t, workerId)
					l.Trace().Func(func(e *zerolog.Event) {
						t.updateMu.RLock()
						defer t.updateMu.RUnlock()
						e.Dur("task_duration_ms", t.ExeTime()).Str("exit_status", string(t.exitStatus)).Msg("Task finished")
					})

					// Dec tasks being processed
					wp.busyCount.Add(-1)

					// Tasks must set their completed flag before exiting
					// if it wasn't done in the work body, we do it for them
					if t.QueueState() != Exited {
						t.Success("closed by worker pool")
					}

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

					wp.cleanup(t, workerId)

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
func (wp *WorkerPool) addReplacementWorker(ctx context.Context) {
	wp.maxWorkers.Add(1)

	log.FromContext(ctx).Debug().Msg("Adding replacement worker")

	wp.execWorker(ctx, true)
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

	ticker := time.NewTicker(time.Second * 10)
	defer func() {
		wp.ctx.Log().Debug().Msg("status reporter exiting")
	}()

	for {
		select {
		case _, ok := <-wp.ctx.Done():
			if !ok {
				return
			}
		case <-ticker.C:
			remaining, total, busy, alive, retrySize := wp.Status()
			if lastCount != remaining || busy != 0 {
				lastCount = remaining
				wp.ctx.Log().Debug().Int("queue_remaining", remaining).Int("queue_total", total).Int("queue_buffered", retrySize).Int("busy_workers", busy).Int("alive_workers", alive).Msg(
					"Worker pool status",
				)
			}
		}
	}
}

func (wp *WorkerPool) bufferDrainer() {
	ticker := time.NewTicker(time.Second * 10)
	defer func() {
		wp.ctx.Log().Debug().Msg("buffer drainer exiting")
	}()

	for {
		select {
		case _, ok := <-wp.ctx.Done():
			if !ok {
				return
			}
			wp.ctx.Log().Warn().Msg("buffer drainer not exiting?")
		case <-ticker.C:
			wp.taskBufferMu.Lock()
			if len(wp.retryBuffer) != 0 && len(wp.taskStream) == 0 {
				for _, t := range wp.retryBuffer {
					wp.taskStream <- t
				}
				wp.retryBuffer = []*Task{}
			}
			wp.taskBufferMu.Unlock()
		}
	}
}

func (wp *WorkerPool) addToRetryBuffer(tasks ...*Task) {
	wp.taskBufferMu.Lock()
	defer wp.taskBufferMu.Unlock()
	wp.retryBuffer = append(wp.retryBuffer, tasks...)
}
