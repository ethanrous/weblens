package task

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/pkg/errors"
)

type TaskPool struct {
	id TaskId

	treatAsGlobal  bool
	hasQueueThread bool

	tasks map[TaskId]*Task
	taskLock sync.Mutex

	totalTasks     atomic.Int64
	completedTasks atomic.Int64
	waiterCount    atomic.Int32
	waiterGate     sync.Mutex
	exitLock       sync.Mutex

	workerPool     *WorkerPool
	parentTaskPool *TaskPool
	createdBy      *Task
	createdAt      time.Time

	cleanupFn func(pool Pool)

	allQueuedFlag atomic.Bool

	erroredTasks chan *Task
}

func (tp *TaskPool) IsRoot() bool {
	if tp == nil {
		return false
	}
	return tp.parentTaskPool == nil || tp.parentTaskPool.IsGlobal()
}

func (tp *TaskPool) GetWorkerPool() *WorkerPool {
	return tp.workerPool
}

func (tp *TaskPool) ID() TaskId {
	return tp.id
}

func (tp *TaskPool) addTask(task *Task) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()
	tp.tasks[task.taskId] = task
}

// // newTask passes params to create new task, and return the task to the caller.
// // If the task already exists, the existing task will be returned, and a new one will not be created
// func (tp *TaskPool) newTask(
// 	jobName weblens.TaskType, taskMeta TaskMetadata, caster websocket.BroadcasterAgent,
// ) *Task {
//
// 	var taskId TaskId
// 	if taskMeta == nil {
// 		taskId = TaskId(internal.GlobbyHash(8, time.Now().String()))
// 	} else {
// 		taskId = TaskId(internal.GlobbyHash(8, taskMeta.MetaString(), jobName))
// 	}
//
// 	existingTask := tp.GetWorkerPool().GetTask(taskId)
// 	if existingTask != nil {
// 		if jobName == "write_file" {
// 			existingTask.ClearAndRecompute()
// 		}
// 		return existingTask
// 	}
//
// 	newTask := &Task{
// 		taskId:    taskId,
// 		jobName:  jobName,
// 		metadata:  taskMeta,
// 		updateMu:  &sync.RWMutex{},
// 		waitMu:    sync.Mutex{},
// 		timerLock: sync.Mutex{},
// 		signal: atomic.Int64{},
//
// 		queueState: PreQueued,
//
// 		// signal chan must be buffered so caller doesn't block trying to close many tasks
// 		signalChan: make(chan int, 1),
//
// 		sw: internal.NewStopwatch(fmt.Sprintf("%s Task [%s]", jobName, taskId)),
// 		caster: caster,
// 	}
//
// 	// Lock the waiter gate immediately. The task cleanup routine will clear
// 	// this lock when the task exits, which will allow any thread waiting on
// 	// the task to return
// 	newTask.waitMu.Lock()
//
// 	switch newTask.jobName {
// 	case weblens.ScanDirectoryTask:
// 		newTask.work = scanDirectory
// 	case weblens.CreateZipTask:
// 		// don't remove task when finished since we can just return the name of the already made zip
// 		// file if asked for the same files again
// 		newTask.persistent = true
// 		newTask.work = weblens.createZipFromPaths
// 	case weblens.ScanFileTask:
// 		newTask.work = weblens.scanFile
// 	case weblens.MoveFileTask:
// 		newTask.work = weblens.moveFile
// 	case weblens.WriteFileTask:
// 		newTask.work = weblens.handleFileUploads
// 	case weblens.GatherFsStatsTask:
// 		newTask.work = weblens.gatherFilesystemStats
// 	case weblens.BackupTask:
// 		newTask.work = doBackup
// 	case weblens.HashFile:
// 		newTask.work = weblens.hashFile
// 	case weblens.CopyFileFromCore:
// 		newTask.work = copyFileFromCore
// 	default:
// 		wlog.ShowErr(error2.WErrMsg("unknown task type"))
// 	}
//
// 	tp.workerPool.addTask(newTask)
//
// 	tp.taskLock.Lock()
// 	defer tp.taskLock.Unlock()
// 	tp.tasks[newTask.taskId] = newTask
//
// 	return newTask
// }

func (tp *TaskPool) RemoveTask(taskId TaskId) {
	tp.taskLock.Lock()
	defer tp.taskLock.Unlock()
	delete(tp.tasks, taskId)

}

// func (tp *TaskPool) ScanDirectory(directory *fileTree.WeblensFile, caster websocket.BroadcasterAgent) *Task {
// 	// Partial media means nothing for a directory scan, so it's always nil
//
// 	if caster == nil {
// 		wlog.Error.Println("caster is nil")
// 		return nil
// 	}
//
// 	scanMeta := weblens.scanMetadata{file: directory}
// 	t := tp.newTask(weblens.ScanDirectoryTask, scanMeta, caster)
// 	err := tp.QueueTask(t)
// 	if err != nil {
// 		wlog.ErrTrace(err)
// 		return nil
// 	}
//
// 	return t
// }
//
// func (tp *TaskPool) ScanFile(file *fileTree.WeblensFile, caster websocket.BroadcasterAgent) *Task {
// 	scanMeta := weblens.scanMetadata{file: file}
// 	t := tp.newTask(weblens.ScanFileTask, scanMeta, caster)
// 	err := tp.QueueTask(t)
// 	if err != nil {
// 		wlog.ErrTrace(err)
// 		return nil
// 	}
//
// 	return t
// }
//
// func (tp *TaskPool) WriteToFile(
// 	rootFolderId types.FileId, chunkSize, totalUploadSize int64,
// 	caster websocket.BroadcasterAgent,
// ) *Task {
// 	numChunks := totalUploadSize / chunkSize
// 	numChunks = int64(math.Max(float64(numChunks), 10))
// 	writeMeta := weblens.writeFileMeta{
// 		rootFolderId: rootFolderId, chunkSize: chunkSize, totalSize: totalUploadSize,
// 		chunkStream: make(chan weblens.fileChunk, numChunks),
// 	}
// 	t := tp.newTask(weblens.WriteFileTask, writeMeta, caster)
//
// 	// We don't queue upload tasks right away, once the first chunk comes through,
// 	// we will add it to the buffer, and then load the task onto the queue.
// 	t.(*Task).setTaskPoolInternal(tp)
//
// 	return t
// }
//
// func (tp *TaskPool) MoveFile(
// 	fileId, destinationFolderId types.FileId, newFilename string, fileEvent types.FileEvent,
// 	caster websocket.BroadcasterAgent,
// ) *Task {
// 	moveMeta := weblens.moveMeta{
// 		fileId: fileId, destinationFolderId: destinationFolderId, newFilename: newFilename,
// 		fileEvent: fileEvent,
// 	}
// 	t := tp.newTask(weblens.MoveFileTask, moveMeta, caster)
// 	err := tp.QueueTask(t)
// 	if err != nil {
// 		wlog.ErrTrace(err)
// 		return nil
// 	}
//
// 	return t
// }
//
// func (tp *TaskPool) CreateZip(
// 	files []*fileTree.WeblensFile, username weblens.Username, shareId types.ShareId,
// 	casters websocket.BroadcasterAgent,
// ) *Task {
// 	meta := weblens.zipMetadata{files: files, username: username, shareId: shareId}
// 	t := tp.newTask(weblens.CreateZipTask, meta, casters)
// 	if c, _ := t.Status(); !c {
// 		err := tp.QueueTask(t)
// 		if err != nil {
// 			wlog.ErrTrace(err)
// 			return nil
// 		}
// 	}
//
// 	return t
// }
//
// func (tp *TaskPool) GatherFsStats(rootDir *fileTree.WeblensFile, caster websocket.BroadcasterAgent) *Task {
// 	t := tp.newTask(weblens.GatherFsStatsTask, weblens.fsStatMeta{rootDir: rootDir}, caster)
// 	err := tp.QueueTask(t)
// 	if err != nil {
// 		wlog.ErrTrace(err)
// 		return nil
// 	}
//
// 	return t
// }
//
// func (tp *TaskPool) Backup(remoteId types.InstanceId, caster websocket.BroadcasterAgent) *Task {
// 	t := tp.newTask(weblens.BackupTask, weblens.backupMeta{remoteId: remoteId}, caster)
// 	err := tp.QueueTask(t)
// 	if err != nil {
// 		wlog.ErrTrace(err)
// 		return nil
// 	}
//
// 	return t
// }
//
// func (tp *TaskPool) CopyFileFromCore(
// 	file *fileTree.WeblensFile, core types.Client, caster websocket.BroadcasterAgent,
// ) *Task {
// 	t := tp.newTask(weblens.CopyFileFromCore, weblens.backupCoreFileMeta{file: file, core: core}, caster)
// 	err := tp.QueueTask(t)
// 	if err != nil {
// 		wlog.ErrTrace(err)
// 		return nil
// 	}
//
// 	return t
// }
//
// func (tp *TaskPool) HashFile(f *fileTree.WeblensFile, caster websocket.BroadcasterAgent) *Task {
// 	t := tp.newTask(weblens.HashFile, weblens.hashFileMeta{file: f}, caster)
// 	err := tp.QueueTask(t)
// 	if err != nil {
// 		wlog.ErrTrace(err)
// 		return nil
// 	}
//
// 	return t
// }

func (tp *TaskPool) handleTaskExit(replacementThread bool) (canContinue bool) {

	tp.completedTasks.Add(1)

	// Global queues do not finish and cannot be waited on. If this is NOT a global queue,
	// we check if we are empty and finished, and if so, wake any waiters.
	if !tp.treatAsGlobal {
		uncompletedTasks := tp.totalTasks.Load() - tp.completedTasks.Load()

		// Even if we are out of tasks, if we have not been told all tasks
		// were queued, we do not wake the waiters
		if uncompletedTasks == 0 && tp.allQueuedFlag.Load() {
			if tp.waiterCount.Load() != 0 {
				log.Debug.Println("Pool complete, waking sleepers!")
				// TODO - move pool completion to cleanup function
				// if tp.createdBy != nil && tp.createdBy.caster != nil {
				// 	tp.createdBy.caster.PushPoolUpdate(tp, websocket.PoolCompleteEvent, nil)
				// }
				tp.waiterGate.Unlock()

				// Check if all waiters have awoken before closing the queue, spin and sleep for 10ms if not
				// Should only loop a handful of times, if at all, we just need to wait for the waiters to
				// lock and then release immediately. Should take, like, nanoseconds (?) each
				for tp.waiterCount.Load() != 0 {
					time.Sleep(time.Millisecond * 10)
				}
			}

			// Once all the tasks have exited, this worker pool is now closing, and so we must run
			// its cleanup routine(s)
			if tp.cleanupFn != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.ShowErr(errors.New(fmt.Sprint(r)), "Failed to execute taskPool cleanup")
						}
					}()
					tp.cleanupFn(tp)
				}()
			}
			tp.workerPool.removeTaskPool(tp.ID())
		}
	}
	// If this is a replacement task, and we have more workers than the target for the pool, we exit
	if replacementThread && tp.workerPool.currentWorkers.Load() > tp.workerPool.maxWorkers.Load() {
		// Important to decrement alive workers inside the exitLock, so
		// we don't have multiple workers exiting when we only need the 1
		tp.workerPool.currentWorkers.Add(-1)

		return false
	}

	// If we have already began running the task,
	// we must finish and clean up before checking exitFlag again here.
	// The task *could* be cancelled to speed things up, but that
	// is not our job.
	if tp.workerPool.exitFlag.Load() == 1 {
		// Dec alive workers
		tp.workerPool.currentWorkers.Add(-1)
		return false
	}

	return true
}

func (tp *TaskPool) GetRootPool() *TaskPool {
	if tp.IsRoot() {
		return tp
	}

	tmpTp := tp
	for !tmpTp.parentTaskPool.IsRoot() {
		tmpTp = tmpTp.parentTaskPool
	}
	return tmpTp
}

func (tp *TaskPool) Status() PoolStatus {
	complete := tp.completedTasks.Load()
	total := tp.totalTasks.Load()
	progress := (float64(complete * 100)) / float64(total)
	if math.IsNaN(progress) {
		progress = 0
	}

	return PoolStatus{
		Complete: complete,
		Total:    total,
		Progress: progress,
		Runtime:  time.Since(tp.createdAt),
	}
}

// Wait Parks the thread on the work queue until all the tasks have been queued and finish.
// **If you never call tp.SignalAllQueued(), the waiters will never wake up**
// Make sure that you SignalAllQueued before parking here if it is the only thread
// loading tasks
func (tp *TaskPool) Wait(supplementWorker bool) {
	// Waiting on global queues does not make sense, they are not meant to end
	// or
	// All the tasks were queued, and they have all finished,
	// so no need to wait, we can "wake up" instantly.
	if tp.treatAsGlobal || (tp.allQueuedFlag.Load() && tp.totalTasks.Load()-tp.completedTasks.Load() == 0) {
		return
	}

	// If we want to park another thread that is currently executing a task,
	// e.g a directory scan waiting for the child file scans to complete,
	// we want to add a worker to the pool temporarily to supplement this one
	if supplementWorker {
		tp.workerPool.busyCount.Add(-1)
		tp.workerPool.addReplacementWorker()
	}

	_, file, line, _ := runtime.Caller(1)
	log.Debug.Printf("Parking on worker pool from %s:%d\n", file, line)

	tp.waiterCount.Add(1)
	tp.waiterGate.Lock()
	//lint:ignore SA2001 We want to wake up when the task is finished, and then signal other waiters to do the same
	tp.waiterGate.Unlock()
	tp.waiterCount.Add(-1)

	log.Debug.Printf("Woke up, returning to %s:%d\n", file, line)

	if supplementWorker {
		tp.workerPool.busyCount.Add(1)
		tp.workerPool.removeWorker()
	}
}

func (tp *TaskPool) LockExit() {
	tp.exitLock.Lock()
}

func (tp *TaskPool) UnlockExit() {
	tp.exitLock.Unlock()
}

func (tp *TaskPool) AddError(t *Task) {
	tp.erroredTasks <- t
}

func (tp *TaskPool) AddCleanup(fn func(pool Pool)) {
	tp.exitLock.Lock()
	defer tp.exitLock.Unlock()
	if tp.allQueuedFlag.Load() && tp.completedTasks.Load() == tp.totalTasks.Load() {
		// Caller expects `AddCleanup` to execute asynchronously, so we must run the
		// cleanup function in its own go routine
		go fn(tp)
		return
	}

	tp.cleanupFn = fn
}

func (tp *TaskPool) Errors() []*Task {
	erroredTasks := []*Task{}
	for len(tp.erroredTasks) != 0 {
		erroredTasks = append(erroredTasks, <-tp.erroredTasks)
	}
	return internal.SliceConvert[*Task](erroredTasks)
}

func (tp *TaskPool) Cancel() {
	tp.allQueuedFlag.Store(true)

	// Dont allow more tasks to join the queue while we are cancelling them
	tp.taskLock.Lock()

	for _, t := range tp.tasks {
		t.Cancel()
		t.Wait()
	}
	tp.taskLock.Unlock()

	// TODO - move this to the cleanup function as well
	// Signal to the client that this pool has been canceled, so we can reflect
	// that in the UI
	// Caster.PushPoolUpdate(tp, websocket.PoolCancelledEvent, nil)

}

func (tp *TaskPool) QueueTask(t *Task) (err error) {

	if tp.workerPool.exitFlag.Load() == 1 {
		log.Warning.Println("Not queuing task while worker pool is going down")
		return
	}

	if t.err != nil {
		// Tasks that have failed will not be re-tried. If the errored task is removed from the
		// task map, then it will be re-tried because the previous error was lost. This can be
		// sometimes be useful, some tasks auto-remove themselves after they finish.
		log.Warning.Println("Not re-queueing task that has error set, please restart weblens to try again")
		return
	}

	if t.taskPool != nil && (t.taskPool != tp || t.queueState != PreQueued) {
		// Task is already queued, we are not allowed to move it to another queue.
		// We can call .ClearAndRecompute() on the task and it will queue it
		// again, but it cannot be transferred
		if t.taskPool != tp {
			log.Warning.Printf("Attempted to re-queue a [%s] task that is already in a queue", t.jobName)
			return
		}
		t.taskPool.tasks[t.taskId] = t
		return
	}

	if tp.allQueuedFlag.Load() {
		// We cannot add tasks to a queue that has been closed
		return werror.WithStack(errors.New("attempting to add task to closed task queue"))
	}

	tp.totalTasks.Add(1)

	if tp.parentTaskPool != nil {
		tmpTp := tp
		for tmpTp.parentTaskPool != nil {
			tmpTp = tmpTp.parentTaskPool
		}
		if tmpTp != tp {
			tmpTp.totalTasks.Add(1)
		}
	}

	// Set the tasks queue
	t.taskPool = tp

	tp.workerPool.lifetimeQueuedCount.Add(1)

	// Put the task in the queue
	t.queueState = InQueue
	if len(tp.workerPool.retryBuffer) != 0 || len(tp.workerPool.taskStream) == cap(tp.workerPool.taskStream) {
		tp.workerPool.addToRetryBuffer(t)
	} else {
		tp.workerPool.taskStream <- t
	}

	t.taskPool.tasks[t.taskId] = t
	return
}

// MarkGlobal specifies the work queue as being a "global" one
func (tp *TaskPool) MarkGlobal() {
	tp.treatAsGlobal = true
}

func (tp *TaskPool) IsGlobal() bool {
	return tp.treatAsGlobal
}

func (tp *TaskPool) CreatedInTask() *Task {
	return tp.createdBy
}

func (tp *TaskPool) SignalAllQueued() {
	if tp.treatAsGlobal {
		log.Error.Println("Attempt to signal all queued for global queue")
	}

	tp.exitLock.Lock()
	// If all tasks finish (e.g. early failure, etc.) before we signal that they are all queued,
	// the final exiting task will not let the waiters out, so we must do it here. We must also
	// remove the task pool from the worker pool for the same reason
	if tp.completedTasks.Load() == tp.totalTasks.Load() {
		tp.waiterGate.Unlock()
		tp.workerPool.removeTaskPool(tp.ID())
	}
	tp.allQueuedFlag.Store(true)
	tp.exitLock.Unlock()

	if tp.hasQueueThread {
		tp.workerPool.removeWorker()
		tp.hasQueueThread = false
	}
}

type PoolStatus struct {
	// The count of tasks that have completed on this task pool
	Complete int64

	// The count of all tasks that have been queued on this task pool
	Total int64

	// Percent to completion of all tasks
	Progress float64

	// How long the pool has been alive
	Runtime time.Duration
}

type Pool interface {
	ID() TaskId

	QueueTask(*Task) error

	MarkGlobal()
	IsGlobal() bool
	SignalAllQueued()

	CreatedInTask() *Task

	Wait(bool)
	Cancel()

	IsRoot() bool
	Status() PoolStatus
	AddCleanup(fn func(Pool))

	GetRootPool() *TaskPool
	GetWorkerPool() *WorkerPool

	LockExit()
	UnlockExit()

	RemoveTask(TaskId)

	// NotifyTaskComplete(Task, websocket.BroadcasterAgent, ...any)

	// ScanDirectory(WeblensFile, websocket.BroadcasterAgent) Task
	// ScanFile(WeblensFile, websocket.BroadcasterAgent) Task
	// WriteToFile(FileId, int64, int64, websocket.BroadcasterAgent) Task
	// MoveFile(FileId, FileId, string, FileEvent, websocket.BroadcasterAgent) Task
	// GatherFsStats(WeblensFile, websocket.BroadcasterAgent) Task
	// Backup(InstanceId, websocket.BroadcasterAgent) Task
	// HashFile(WeblensFile, websocket.BroadcasterAgent) Task
	// CreateZip(files []WeblensFile, username Username, shareId ShareId, casters websocket.BroadcasterAgent) Task
	// CopyFileFromCore(WeblensFile, Client, websocket.BroadcasterAgent) Task

	Errors() []*Task
	AddError(t *Task)
}
