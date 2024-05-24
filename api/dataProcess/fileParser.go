package dataProcess

import (
	"errors"
	"fmt"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

// Global exiftool
var gexift *exiftool.Exiftool
var gexiftBufferSize int64

func InitGExif(bufSize int64) *exiftool.Exiftool {
	if bufSize <= gexiftBufferSize {
		return gexift
	}
	if gexift != nil {
		err := gexift.Close()
		util.ErrTrace(err)
		gexift = nil
		gexiftBufferSize = 0
	}
	buf := make([]byte, int(bufSize))
	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"), exiftool.ExtractAllBinaryMetadata(), exiftool.Buffer(buf, int(bufSize)))
	if err != nil {
		util.ErrTrace(err)
		return nil
	}
	gexift = et

	return gexift
}

func processMediaFile(t *task) {
	meta := t.metadata.(ScanMetadata)
	m := meta.partialMedia
	file := meta.file

	file.AddTask(t)
	defer file.RemoveTask(t.TaskId())

	if m == nil {
		t.ErrorAndExit(errors.New("attempted to process nil media"))
		return
	}

	if !file.IsDisplayable() {
		return
	}

	m, err := m.LoadFromFile(file, meta.fileBytes, t)
	if err != nil {
		t.ErrorAndExit(err)
		return
	}

	t.CheckExit()

	err = m.Save()
	if err != nil {
		t.ErrorAndExit(err)
		return
	}

	t.CheckExit()

	m.Clean()
	if !m.IsCached() {
		t.ErrorAndExit(fmt.Errorf("media scan exiting due to missing cache"))
	}
	m.SetImported(true)

	if t.caster != nil {
		t.caster.PushFileUpdate(file)
		t.taskPool.NotifyTaskComplete(t, t.caster)
	}

	t.success()
}

func scanDirectory(t *task) {
	meta := t.metadata.(ScanMetadata)
	scanDir := meta.file

	if scanDir.Filename() == ".user_trash" {
		t.taskPool.NotifyTaskComplete(t, t.caster, "No media to scan")
		globalCaster.PushTaskUpdate(t.taskId, ScanComplete, types.TaskResult{"execution_time": t.ExeTime()}) // Let any client subscribers know we are done
		t.success("No media to scan")
		return
	}

	// "deepScan" defines wether or not we should go sync
	// with the real filesystem for this scan. Otherwise,
	// just handle processing media we already know about
	if meta.deepScan {
		scanDir.ReadDir()
	}

	// Claim task lock on this file before reading. This
	// prevents lost scans on child files if we were, say,
	// uploading into this directory as a scan comes through.
	// We will block until the upload finishes, then continue this scan
	scanDir.AddTask(t)
	defer scanDir.RemoveTask(t.TaskId())

	tp := NewTaskPool(true, t)
	util.Info.Printf("Beginning directory scan for %s\n", scanDir.GetAbsPath())

	t.caster.FolderSubToTask(scanDir.Id(), t.TaskId())
	t.caster.FolderSubToTask(scanDir.GetParent().Id(), t.TaskId())

	err := scanDir.LeafMap(func(wf types.WeblensFile) error {
		if wf.IsDir() {
			return nil
			// TODO: Lock directory files while scanning to be able to check what task is using each file
			wf.AddTask(t)
		}
		// If this file is already being processed, don't queue it again
		fileTask := wf.GetTask()
		if fileTask != nil && fileTask.TaskType() == ScanFileTask {
			return nil
		}

		if !wf.IsDisplayable() {
			return nil
		}

		m := dataStore.MediaMapGet(wf.GetContentId())
		if m != nil && m.IsImported() && m.IsCached() {
			return nil
		}

		tp.ScanFile(wf, t.caster)
		return nil
	})

	if err != nil {
		t.ErrorAndExit(err)
	}

	tp.SignalAllQueued()
	tp.Wait(true)

	errs := tp.Errors()
	if len(errs) != 0 {
		globalCaster.PushTaskUpdate(t.taskId, TaskFailed, types.TaskResult{"failure_note": fmt.Sprintf("%d scans failed", len(errs))}) // Let any client subscribers know we are done
		t.ErrorAndExit(ErrChildTaskFailed)
	}

	globalCaster.PushTaskUpdate(t.taskId, ScanComplete, types.TaskResult{"execution_time": t.ExeTime()}) // Let any client subscribers know we are done
	tp.NotifyTaskComplete(t, t.caster)
	t.success()
}

func getScanResult(t *task) types.TaskResult {
	var tp *taskPool

	if t.taskPool != nil {
		tp = t.taskPool.GetRootPool()
	}

	result := types.TaskResult{
		"filename": t.metadata.(ScanMetadata).file.Filename(),
	}

	if tp != nil {
		complete, total, progress := tp.status()
		result["percent_progress"] = progress
		result["tasks_complete"] = complete
		result["tasks_total"] = total
		result["task_job_name"] = tp.createdBy.TaskType()
		result["task_job_target"] = tp.createdBy.metadata.(ScanMetadata).file.Filename()
	} else {
		result["task_job_name"] = t.TaskType()
		result["task_job_target"] = t.metadata.(ScanMetadata).file.Filename()
	}

	return result
}
