package dataProcess

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func (tp *taskPool) ScanDirectory(directory types.WeblensFile, recursive, deep bool, caster types.BroadcasterAgent) types.Task {
	// Partial media means nothing for a directory scan, so it's always nil
	scanMeta := ScanMetadata{file: directory, recursive: recursive, deepScan: deep}
	t := newTask(ScanDirectoryTask, scanMeta, caster, nil)
	tp.QueueTask(t)

	if caster != nil {
		caster.PushTaskUpdate(t.TaskId(), TaskCreated, types.TaskResult{
			"taskType":      ScanDirectoryTask,
			"directoryName": directory.Filename(),
		})
	}

	return t
}

func (tp *taskPool) ScanFile(file types.WeblensFile, fileBytes []byte, caster types.BroadcasterAgent) types.Task {
	scanMeta := ScanMetadata{file: file, fileBytes: fileBytes}
	t := newTask(ScanFileTask, scanMeta, caster, nil)
	tp.QueueTask(t)

	return t
}

// Parameters:
//   - `filename` : the name of the new file to create
//   - `fileSize` : the parent directory to upload the new file into
func (tp *taskPool) WriteToFile(rootFolderId types.FileId, chunkSize, totalUploadSize int64, caster types.BroadcasterAgent) types.Task {
	numChunks := totalUploadSize / chunkSize
	numChunks = int64(math.Max(float64(numChunks), 10))
	writeMeta := WriteFileMeta{rootFolderId: rootFolderId, chunkSize: chunkSize, totalSize: totalUploadSize, chunkStream: make(chan FileChunk, numChunks)}
	t := newTask(WriteFileTask, writeMeta, caster, nil)

	// We don't queue upload tasks right away, once the first chunk comes through,
	// we will add it to the buffer, and then load the task onto the queue

	// if !t.completed {
	// 	tskr.QueueTask(t)
	// }

	return t
}

func (tp *taskPool) MoveFile(fileId, destinationFolderId types.FileId, newFilename string, caster types.BroadcasterAgent) types.Task {
	moveMeta := MoveMeta{fileId: fileId, destinationFolderId: destinationFolderId, newFilename: newFilename}
	t := newTask(MoveFileTask, moveMeta, caster, nil)
	tp.QueueTask(t)

	return t
}

func (tp *taskPool) CreateZip(files []types.WeblensFile, username types.Username, shareId types.ShareId, casters types.BroadcasterAgent) types.Task {
	meta := ZipMetadata{files: files, username: username, shareId: shareId}
	t := newTask(CreateZipTask, meta, casters, nil)
	if c, _ := t.Status(); !c {
		tp.QueueTask(t)
	}

	return t
}

func (tp *taskPool) GatherFsStats(rootDir types.WeblensFile, caster types.BroadcasterAgent) types.Task {
	t := newTask(GatherFsStatsTask, FsStatMeta{rootDir: rootDir}, caster, nil)
	tp.QueueTask(t)

	return t
}

func (tp *taskPool) Backup(remoteId string, requester types.Requester) types.Task {
	t := newTask(BackupTask, BackupMeta{remoteId: remoteId}, nil, requester)
	tp.QueueTask(t)

	return t
}

func (tp *taskPool) HashFile(f types.WeblensFile) types.Task {
	t := newTask(HashFile, HashFileMeta{file: f}, nil, nil)
	tp.QueueTask(t)

	return t
}

func (tp *taskPool) handleTaskExit(replacementThread bool) (canContinue bool) {

	tp.completedTasks.Add(1)

	// Global queues do not finish and cannot be waited on. If this is NOT a global queue,
	// we check if we are empty and finished, and if so wake any waiters.
	if !tp.treatAsGlobal {
		uncompletedTasks := tp.totalTasks.Load() - tp.completedTasks.Load()
		// util.Debug.Println("Uncompleted tasks:", uncompletedTasks)
		// util.Debug.Println("Waiter count:", tp.waiterCount.Load())
		// util.Debug.Println("All queued:", tp.allQueuedFlag)

		// Even if we are out of tasks, if we have not been told all tasks
		// were queued, we do not wake the waiters
		if uncompletedTasks == 0 && tp.waiterCount.Load() != 0 && tp.allQueuedFlag {
			util.Debug.Println("Waking sleepers!")
			tp.waiterGate.Unlock()

			// Check if all waiters have awoken before closing the queue, spin and sleep for 10ms if not
			// Should only loop a handful of times, if at all, we just need to wait for the waiters to
			// lock and then release immediately, should take nanoseconds each
			for tp.waiterCount.Load() != 0 {
				time.Sleep(time.Millisecond * 10)
			}
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
	if tp.workerPool.exitFlag == 1 {
		// Dec alive workers
		tp.workerPool.currentWorkers.Add(-1)
		return false
	}

	return true
}

func (tp *taskPool) GetRootPool() *taskPool {
	if tp == nil || tp.treatAsGlobal {
		return nil
	}

	tmpTp := tp
	for tmpTp.parentTaskPool != nil {
		tmpTp = tmpTp.parentTaskPool
	}
	return tmpTp
}

func (tp *taskPool) status() (int, int, float64) {
	complete := tp.completedTasks.Load() + 1
	total := tp.totalTasks.Load()
	progress := (float64(complete * 100)) / float64(total)

	return int(complete), int(total), progress
}

func (tp *taskPool) ClearAllQueued() {
	if tp.waiterCount.Load() != 0 {
		util.Warning.Println("Clearing all queued flag on work queue that still has sleepers")
	}
	tp.allQueuedFlag = false
}

func (tp *taskPool) NotifyTaskComplete(t *task, c types.BroadcasterAgent, note ...any) {
	rp := t.taskPool.GetRootPool()
	var rootTask *task
	if rp != nil && rp.createdBy != nil {
		rootTask = rp.createdBy
	} else {
		rootTask = t
	}

	var result types.TaskResult
	switch t.taskType {
	case ScanDirectoryTask, ScanFileTask:
		result = getScanResult(t)
	default:
		return
	}

	if len(note) != 0 {
		result["note"] = fmt.Sprint(note...)
	}

	c.PushTaskUpdate(rootTask.TaskId(), SubTaskComplete, result)

}

// Park the thread on the work queue until all the tasks have been queued and finish.
// **If you never call tp.SignalAllQueued(), the waiters will never wake up**
// Make sure that you SignalAllQueued before parking here if it is the only thread
// loading tasks
func (tp *taskPool) Wait(supplementWorker bool) {
	// Waiting on global queues does not make sense, they are not meant to end
	// or
	// All the tasks were queued, and they have all finished,
	// so no need to wait, we can "wake up" instantly.
	if tp.treatAsGlobal || (tp.allQueuedFlag && tp.totalTasks.Load()-tp.completedTasks.Load() == 0) {
		return
	}

	// If we want to park another thread that is currently executing a task,
	// e.g a directory scan waiting for the child file scans to complete,
	// we want to add an additional worker to the pool temporarily to supplement this one
	if supplementWorker {
		tp.workerPool.addReplacementWorker()
	}

	_, file, line, _ := runtime.Caller(1)
	util.Debug.Printf("Parking on worker pool from %s:%d\n", file, line)

	tp.workerPool.busyCount.Add(-1)
	tp.waiterCount.Add(1)
	tp.waiterGate.Lock()
	//lint:ignore SA2001 We want to wake up when the task is finished, and then signal other waiters to do the same
	tp.waiterGate.Unlock()
	tp.waiterCount.Add(-1)
	tp.workerPool.busyCount.Add(1)

	util.Debug.Printf("Woke up, returning to %s:%d\n", file, line)

	if supplementWorker {
		tp.workerPool.removeWorker()
	}
}
