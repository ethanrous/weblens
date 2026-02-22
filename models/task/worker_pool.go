package task

import (
	"context"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethanrous/weblens/modules/log"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// ErrTaskError indicates a task encountered an error during execution.
var ErrTaskError = wlerrors.New("task error")

// ErrTaskExit indicates a task has exited.
var ErrTaskExit = wlerrors.New("task exit")

// ErrTaskTimeout indicates a task exceeded its timeout duration.
var ErrTaskTimeout = wlerrors.New("task timeout")

// ErrTaskCanceled indicates a task was canceled before completion.
var ErrTaskCanceled = wlerrors.New("task canceled")

// ErrTaskAlreadyComplete indicates an attempt to execute an already completed task.
var ErrTaskAlreadyComplete = wlerrors.New("task already complete")

type hit struct {
	time   time.Time
	target *Task
}

type workChannel chan *Task
type hitChannel chan hit

type job struct {
	handler HandlerFunc
	opts    Options
}

// HandlerFunc defines the function signature for task execution functions.
type HandlerFunc func(task *Task)

// WorkerPool manages a pool of workers that execute tasks from task pools.
type WorkerPool struct {
	ctx context.Context

	busyCount *atomic.Int64 // Number of workers currently executing a task

	registeredJobs map[string]job

	taskMap map[string]*Task

	taskPoolMap map[string]*Pool

	taskStream workChannel
	hitStream  hitChannel

	retryBuffer    []*Task
	maxWorkers     atomic.Int64 // Max allowed worker count
	currentWorkers atomic.Int64 // Current total of workers on the pool

	lifetimeQueuedCount atomic.Int64

	jobsMu sync.RWMutex

	taskMu sync.RWMutex

	taskPoolMu   sync.Mutex
	taskBufferMu sync.Mutex
}

// NewWorkerPool creates and initializes a new worker pool with the specified number of workers.
func NewWorkerPool(initWorkers int) *WorkerPool {
	if initWorkers == 0 {
		initWorkers = 1
	}

	newWp := &WorkerPool{
		registeredJobs: map[string]job{},
		taskMap:        map[string]*Task{},
		taskPoolMap:    map[string]*Pool{},

		busyCount: &atomic.Int64{},

		taskStream:  make(workChannel, initWorkers*1000),
		retryBuffer: []*Task{},

		hitStream: make(hitChannel, initWorkers*2),
	}

	newWp.maxWorkers.Store(int64(initWorkers))

	return newWp
}

// NewTaskPool `replace` spawns a temporary replacement thread on the parent worker pool.
// This prevents a deadlock when the queue fills up while adding many tasks, and none are being executed
//
// `parent` allows chaining of task pools for floating updates to the top. This makes
// it possible for clients to subscribe to a single task, and get notified about
// all of the sub-updates of that task
// See taskPool.go
func (wp *WorkerPool) NewTaskPool(replace bool, createdBy *Task) (*Pool, error) {
	if wp.ctx == nil {
		return nil, wlerrors.New("worker pool not running, cannot create task pool")
	}

	tp := wp.newTaskPoolInternal()

	if createdBy != nil {
		t := createdBy
		tp.createdBy = t

		if !createdBy.GetTaskPool().IsGlobal() {
			tp.createdBy = t
		}
	}

	if replace {
		// We want to use the same context as the worker pool, but we need a cancelable context for when the task pool is closed
		ctx, cancel := context.WithCancel(wp.ctx)

		tp.AddCleanup(func(_ *Pool) { cancel() })
		wp.addReplacementWorker(ctx)

		tp.hasQueueThread = true
	}

	wp.taskPoolMu.Lock()
	defer wp.taskPoolMu.Unlock()

	wp.taskPoolMap[tp.ID()] = tp

	if createdBy != nil {
		createdBy.SetChildTaskPool(tp)
	}

	return tp, nil
}

// Status returns the count of tasks in the queue, the total number of tasks accepted,
// number of busy workers, and the total number of live workers in the worker pool
func (wp *WorkerPool) Status() (int, int, int, int, int) {
	total := int(wp.lifetimeQueuedCount.Load())

	wp.taskBufferMu.Lock()
	defer wp.taskBufferMu.Unlock()

	return len(wp.taskStream), total, int(wp.busyCount.Load()), int(wp.currentWorkers.Load()), len(wp.retryBuffer)
}

// GetTaskPool returns the task pool with the specified ID.
func (wp *WorkerPool) GetTaskPool(tpID string) *Pool {
	wp.taskPoolMu.Lock()
	defer wp.taskPoolMu.Unlock()

	return wp.taskPoolMap[tpID]
}

// GetTasksByJobName returns all tasks with the specified job name.
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

// GetTaskPoolByJobName returns the task pool created by a task with the specified job name.
func (wp *WorkerPool) GetTaskPoolByJobName(jobName string) *Pool {
	wp.taskPoolMu.Lock()
	defer wp.taskPoolMu.Unlock()

	for _, tp := range wp.taskPoolMap {
		if tp.CreatedInTask() != nil && tp.CreatedInTask().JobName() == jobName {
			return tp
		}
	}

	return nil
}

// RegisterJob adds a template for a repeatable job that can be called upon later in the program
func (wp *WorkerPool) RegisterJob(jobName string, fn HandlerFunc, opts ...Options) {
	wp.jobsMu.Lock()
	defer wp.jobsMu.Unlock()

	o := Options{}
	if len(opts) != 0 {
		o = opts[0]
	}

	wp.registeredJobs[jobName] = job{handler: fn, opts: o}
}

// DispatchJob creates and queues a new task for the specified registered job.
func (wp *WorkerPool) DispatchJob(ctx context.Context, jobName string, meta Metadata, pool *Pool) (*Task, error) {
	if meta.JobName() != jobName {
		return nil, wlerrors.Errorf("job name does not match task metadata")
	}

	if err := meta.Verify(); err != nil {
		return nil, err
	}

	wp.jobsMu.RLock()

	if wp.registeredJobs[jobName].handler == nil {
		wp.jobsMu.RUnlock()

		return nil, wlerrors.Errorf("trying to dispatch non-registered job: %s", jobName)
	}

	wp.jobsMu.RUnlock()

	if pool == nil {
		pool = wp.GetTaskPool(GlobalTaskPoolID)
	}

	job := wp.getRegisteredJob(jobName)

	taskID := makeTaskID(meta, job.opts.Unique)

	t := wp.GetTask(taskID)

	if t != nil {
		t.Log().Warn().Msgf("Task [%s] already exists, not re-queuing", taskID)

		return t, nil
	}

	newl := log.FromContext(ctx).With().
		Str("task_id", taskID).
		Str("job_name", jobName).
		Logger()

	ctx = log.WithContext(ctx, &newl)

	ctx, cancel := context.WithCancelCause(ctx)

	t = &Task{
		taskID:   taskID,
		jobName:  jobName,
		metadata: meta,
		work:     job,

		queueState: Created,

		// signal chan must be buffered so caller doesn't block trying to close many tasks
		waitChan: make(chan struct{}),

		Ctx:        ctx,
		cancelFunc: cancel,
	}

	wp.addTask(t)

	t.Log().Trace().Stack().Msgf("Task [%s] created", taskID)

	select {
	case _, ok := <-t.Ctx.Done():
		if !ok {
			return nil, wlerrors.New("not queuing task while worker pool is going down")
		}
	default:
	}

	if t.err != nil {
		// Tasks that have failed will not be re-tried. If the errored task is removed from the
		// task map, then it will be re-tried because the previous error was lost. This can be
		// sometimes be useful, some tasks auto-remove themselves after they finish.
		return nil, wlerrors.New("Not re-queuing task that has error set")
	}

	if t.taskPool != nil && (t.taskPool != pool || t.queueState != Created) {
		// Task is already queued, we are not allowed to move it to another queue.
		// We can call .ClearAndRecompute() on the task and it will queue it
		// again, but it cannot be transferred
		if t.taskPool != pool {
			return nil, wlerrors.Errorf("Attempted to re-queue a [%s] task that is already in another queue", t.jobName)
		}

		pool.addTask(t)

		return t, nil
	}

	if pool.allQueuedFlag.Load() {
		// We cannot add tasks to a queue that has been closed
		return nil, wlerrors.Errorf("attempting to add task [%s] to closed task queue [pool created by %s]", t.JobName(), pool.ID())
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

	t.SetQueueTime(time.Now())

	pool.addTask(t)

	return t, nil
}

// AddHit schedules a timeout check for the specified task at the given time.
func (wp *WorkerPool) AddHit(time time.Time, target *Task) {
	wp.hitStream <- hit{time: time, target: target}
}

// Run launches the standard threads for this worker pool
func (wp *WorkerPool) Run(ctx context.Context) {
	wp.ctx = ctx

	// Worker pool always has one global queue
	globalPool := wp.newTaskPoolInternal()
	globalPool.id = GlobalTaskPoolID
	globalPool.MarkGlobal()

	wp.taskPoolMap[GlobalTaskPoolID] = globalPool

	// Spawn the timeout checker
	go wp.reaper(ctx)

	// Spawn the buffer worker
	go wp.bufferDrainer(ctx)

	// Spawn the status printer
	go wp.statusReporter(ctx)

	for range wp.maxWorkers.Load() {
		// These are the base, 'omnipresent' threads for this pool,
		// so they are NOT replacement workers. See wp.execWorker
		// for more info
		wp.execWorker(ctx, false)
	}
}

// GetTask returns the task with the specified ID.
func (wp *WorkerPool) GetTask(taskID string) *Task {
	wp.taskMu.RLock()
	defer wp.taskMu.RUnlock()

	return wp.taskMap[taskID]
}

// GetTasks returns all tasks currently managed by this worker pool.
func (wp *WorkerPool) GetTasks() []*Task {
	wp.taskMu.RLock()
	defer wp.taskMu.RUnlock()

	return slices.Collect(maps.Values(wp.taskMap))
}

func (wp *WorkerPool) addTask(task *Task) {
	wp.taskMu.Lock()
	defer wp.taskMu.Unlock()

	wp.taskMap[task.ID()] = task
}

func (wp *WorkerPool) removeTask(taskID string) {
	wp.taskMu.RLock()
	t := wp.taskMap[taskID]
	wp.taskMu.RUnlock()

	if t == nil {
		return
	}

	t.GetTaskPool().RemoveTask(taskID)

	wp.taskMu.Lock()
	defer wp.taskMu.Unlock()

	delete(wp.taskMap, taskID)
}

// execWorker is the "main" function for a worker. It spawns a worker and loops over the task channel
// to pick up and execute tasks.
//
// `isReplacement` specifies if the worker is a temporary replacement for another
// task that is parking for a long time. Replacement workers behave slightly
// differently in attempt to minimize parked time of the other task.
func (wp *WorkerPool) execWorker(workerCtx context.Context, isReplacement bool) {
	go func(workerID int64) {
		newl := log.FromContext(workerCtx).With().
			Int("worker_id", int(workerID)).
			Logger()

		workerCtx = log.WithContext(workerCtx, &newl)

		log.FromContext(wp.ctx).Debug().Msgf("Launching worker")

		err := context_mod.AddToWg(workerCtx)
		if err != nil {
			log.FromContext(wp.ctx).Error().Stack().Err(err).Msg("Failed to add worker to wait group")

			return
		}

		defer func() {
			err = context_mod.WgDone(workerCtx)
			if err != nil {
				log.FromContext(wp.ctx).Error().Stack().Err(err).Msg("Failed to remove worker from wait group")

				return
			}

			remainingWorkers := wp.currentWorkers.Add(-1)
			log.FromContext(workerCtx).Debug().Int64("remaining_workers_count", remainingWorkers).Msgf("worker exiting")
		}()

		// WorkLoop:
		for {
			select {
			case _, ok := <-workerCtx.Done():
				if !ok {
					return
				}
			case t := <-wp.taskStream:
				{
					// Check if the task was canceled while in the queue
					select {
					case <-t.Ctx.Done():
						// Task was canceled, do not run it
						log.FromContext(t.Ctx).Trace().Msgf("Task was canceled while in queue, not running")

						continue
					default:
					}

					if t.exitStatus != TaskNoStatus {
						// If the task has already been completed, we don't want to run it again
						log.FromContext(t.Ctx).Trace().Str("exit_status", string(t.exitStatus)).Msgf("Task already has exit status, not running")

						continue
					}

					// Replacement workers are not allowed to do "scan_directory" tasks
					// TODO: - generalize
					if isReplacement && t.jobName == "scan_directory" && t.exitStatus == TaskNoStatus {
						// If there are twice the number of free spaces in the chan, don't bother pulling
						// everything into the waiting buffer, just put it at the end right now.
						if cap(wp.taskStream)-len(wp.taskStream) > int(wp.currentWorkers.Load())*2 {
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
						Int("worker_id", int(workerID)).
						Logger()

					t.setWorkerID(workerID)

					t.ctxMu.Lock()

					// Update the task's logger to include the worker ID
					t.Ctx = log.WithContext(t.Ctx, &newl)

					t.ctxMu.Unlock()

					// Inc tasks being processed
					wp.busyCount.Add(1)
					log.FromContext(t.Ctx).Debug().Func(func(e *zerolog.Event) {
						t.updateMu.RLock()
						defer t.updateMu.RUnlock()

						e.Dur("queue_time_ms", t.QueueTimeDuration()).Msgf("Starting task")
					})

					// Perform the task. This method includes a recover to catch panics in tasks and handle them gracefully.
					// All the real work happens inside safetyWork here, everything before is setup, everything after is teardown.
					wp.safetyWork(t, workerID)

					log.FromContext(t.Ctx).Debug().Func(func(e *zerolog.Event) {
						t.updateMu.RLock()
						defer t.updateMu.RUnlock()

						e.Dur("task_duration_ms", t.ExeTime()).Dur("queue_time_ms", t.QueueTimeDuration()).Str("exit_status", string(t.exitStatus)).Msgf("Task finished")
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
					result["task_id"] = t.taskID
					result["exit_status"] = t.exitStatus

					wp.taskPoolMu.Lock()

					var complete int64

					for _, p := range wp.taskPoolMap {
						status := p.Status()
						complete += status.Complete
					}

					wp.taskPoolMu.Unlock()

					result["queue_remaining"] = complete
					result["queue_total"] = wp.lifetimeQueuedCount.Load()

					t.updateMu.Unlock()

					// Wake any waiters on this task
					close(t.waitChan)

					// Potentially find the task pool that houses this task pool. All child
					// task pools report their status to the root task pool as well.
					// Do not use any global pool as the root
					rootTaskPool := t.taskPool.GetRootPool()

					if t.exitStatus == TaskError && !rootTaskPool.IsGlobal() {
						rootTaskPool.AddError(t)
					}

					// Remove the task from the worker pool's task map if it is not a "persistent" task.
					if !t.work.opts.Persistent {
						wp.removeTask(t.taskID)
					}

					// Update parent task pools about completed task
					directParent := t.GetTaskPool()
					directParent.IncCompletedTasks(1)

					// Run cleanup routines for the task
					wp.cleanupTask(t, workerID)

					var canContinue bool

					if directParent.IsRoot() {
						// Updating the number of workers and then checking it's value is dangerous
						// to do concurrently. Specifically, the waiterGate lock on the queue will,
						// very rarely, attempt to unlock twice if two tasks finish at the same time.
						// So we must treat this whole area as a critical section
						directParent.LockExit()

						// Set values and notifications now that task has completed. Returns
						// a bool that specifies if this thread should continue and grab another
						// task, or if it should exit
						canContinue = directParent.HandleTaskExit(isReplacement)

						directParent.UnlockExit()
					} else {
						rootTaskPool.IncCompletedTasks(1)

						// Must hold both locks (and must acquire root lock first) to enter a dual-update.
						// Any other ordering will result in race conditions or deadlocks
						rootTaskPool.LockExit()

						directParent.LockExit()
						uncompletedTasks := directParent.GetTotalTaskCount() - directParent.GetCompletedTaskCount()
						log.FromContext(wp.ctx).Debug().Msgf(
							"Uncompleted tasks on tp created by %s: %d",
							directParent.CreatedInTask().ID(), uncompletedTasks-1,
						)

						canContinue = directParent.HandleTaskExit(isReplacement)

						// We *should* get the same canContinue value from here, so we do not
						// check it a second time. If we *don't* get the same value, we can safely ignore it
						rootTaskPool.HandleTaskExit(isReplacement)

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

	wp.execWorker(ctx, true)
}

func (wp *WorkerPool) removeWorker() {
	wp.maxWorkers.Add(-1)
}

func (wp *WorkerPool) newTaskPoolInternal() *Pool {
	tpID, err := uuid.NewUUID()
	if err != nil {
		log.FromContext(wp.ctx).Error().Err(err).Msg("Failed to generate UUID for new task pool")

		return nil
	}

	newQueue := &Pool{
		id:           tpID.String(),
		tasks:        map[string]*Task{},
		workerPool:   wp,
		createdAt:    time.Now(),
		erroredTasks: make([]*Task, 0),
		waiterGate:   make(chan struct{}),
		log:          log.FromContext(wp.ctx).With().Str("task_pool", tpID.String()).Logger(),
	}

	return newQueue
}

func (wp *WorkerPool) removeTaskPool(tpID string) {
	wp.taskPoolMu.Lock()
	defer wp.taskPoolMu.Unlock()

	delete(wp.taskPoolMap, tpID)
}

// The reaper handles timed cancelation of tasks. If a task might
// wait on some action from the client, it should queue a timeout with
// the reaper so it doesn't hang forever
func (wp *WorkerPool) reaper(ctx context.Context) {
	timerStream := make(chan *Task)

	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				log.FromContext(ctx).Debug().Msg("Task reaper exiting")

				return
			}

			log.GlobalLogger().Warn().Msg("Reaper not exiting?")
		case newHit := <-wp.hitStream:
			go func(h hit) { time.Sleep(time.Until(h.time)); timerStream <- h.target }(newHit)
		case task := <-timerStream:
			// Possible that the task has kicked its timeout down the road
			// since it first queued it, we must check that we are past the
			// timeout before canceling the task. Also check that it
			// has not already finished
			timeout := task.GetTimeout()
			if task.QueueState() != Exited && time.Until(timeout) <= 0 && timeout.Unix() != 0 {
				log.GlobalLogger().Warn().Msgf("Sending timeout signal to T[%s]\n", task.taskID)
				task.Cancel()
				task.error(ErrTaskTimeout)
			}
		}
	}
}

func (wp *WorkerPool) statusReporter(ctx context.Context) {
	var lastCount int

	ticker := time.NewTicker(time.Second * 10)

	defer func() {
		log.FromContext(ctx).Debug().Msg("status reporter exiting")
	}()

	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				return
			}
		case <-ticker.C:
			remaining, total, busy, alive, retrySize := wp.Status()
			if lastCount != remaining || busy != 0 {
				lastCount = remaining
				log.FromContext(ctx).Debug().Int("queue_remaining", remaining).Int("queue_total", total).Int("queue_buffered", retrySize).Int("busy_workers", busy).Int("alive_workers", alive).Msgf(
					"Worker pool status - %d tasks in queue, %d total tasks queued, %d busy workers, %d alive workers", remaining, total, busy, alive,
				)
			}
		}
	}
}

func (wp *WorkerPool) bufferDrainer(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)

	defer func() {
		log.FromContext(ctx).Debug().Msg("buffer drainer exiting")
	}()

	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				return
			}

			log.FromContext(ctx).Warn().Msg("buffer drainer not exiting?")
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

func (wp *WorkerPool) getRegisteredJob(jobName string) job {
	wp.jobsMu.RLock()
	defer wp.jobsMu.RUnlock()

	return wp.registeredJobs[jobName]
}

func (wp *WorkerPool) workerRecover(task *Task, _ int64) {
	recovered := recover()
	if recovered != nil {
		// Make sure what we got is an error
		var err error

		switch e := recovered.(type) {
		case error:
			if wlerrors.Is(e, ErrTaskError) {
				return
			} else if wlerrors.Is(e, ErrTaskExit) {
				return
			}

			err = wlerrors.WithStack(e)
		default:
			err = wlerrors.Errorf("%s", recovered)
		}

		task.error(err)
	}
}

// saftyWork wraps the task execution with a recover, so if there are any panics
// during the task, we can catch them, display them, and safely remove the task.
func (wp *WorkerPool) safetyWork(task *Task, workerID int64) {
	defer wp.workerRecover(task, workerID)

	task.SetQueueState(Executing)

	if task.exitStatus != TaskNoStatus {
		task.Log().Trace().Msgf("Task [%s] already has exit status [%s], not running", task.taskID, task.exitStatus)
	} else {
		task.updateMu.Lock()
		task.StartTime = time.Now()
		task.updateMu.Unlock()

		task.work.handler(task)

		task.updateMu.Lock()
		task.FinishTime = time.Now()
		task.updateMu.Unlock()
	}
}

// cleanupTask runs the cleanupTask routines for the specified task.
func (wp *WorkerPool) cleanupTask(task *Task, workerID int64) {
	defer wp.workerRecover(task, workerID)

	// Run the cleanup routine for errors, if any
	if task.exitStatus == TaskError || task.exitStatus == TaskCanceled {
		task.Log().Trace().Func(func(e *zerolog.Event) {
			e.Int("error_cleanup_count", len(task.errorCleanups)).Msgf("Task has %d error cleanup(s) to run", len(task.errorCleanups))
		})

		for _, ecf := range task.errorCleanups {
			taskNoCancel(task, ecf)
		}

		task.Log().Trace().Msgf("Finished error cleanup")
	}

	// We want the pool completed and error count to reflect the task has completed
	// when we are doing cleanup. Cleanup is intended to execute "after" the task
	// has finished, so we must inc the completed tasks counter (above) before cleanup
	task.updateMu.RLock()
	cleanups := task.cleanups
	task.updateMu.RUnlock()

	task.Log().Trace().Func(func(e *zerolog.Event) {
		e.Int("cleanup_count", len(cleanups)).Msgf("Task has %d cleanup(s) to run", len(cleanups))
	})

	for _, cf := range cleanups {
		taskNoCancel(task, cf)
	}

	task.Log().Trace().Msgf("Finished cleanups for task [%s]", task.taskID)

	// The post action is intended to run after the task has completed, and after
	// the cleanup has run. This is useful for tasks that need to use the result of
	// the task to do something else, but don't want to block until the task is done
	if task.postAction != nil && task.exitStatus == TaskSuccess {
		task.Log().Trace().Msg("Running task post-action")
		task.postAction(task.result)
		task.Log().Trace().Msg("Finished task post-action")
	}
}

func makeTaskID(meta Metadata, unique bool) string {
	var taskID string

	if meta == nil {
		taskID = globbyHash(8, time.Now().String())
	} else {
		metaStr := meta.MetaString()
		if unique {
			metaStr += time.Now().String()
		}

		taskID = globbyHash(8, metaStr)
	}

	return taskID
}

func taskNoCancel(task *Task, cf HandlerFunc) {
	// We do not want cleanup functions to be cancelable, so we use a context
	// that cannot be canceled. This prevents cleanup functions from being
	// interrupted by task cancelation, which would be bad
	preCtx := task.Ctx
	noCancelCtx := context.WithoutCancel(preCtx)
	task.Ctx = noCancelCtx

	cf(task)

	task.Ctx = preCtx
}
