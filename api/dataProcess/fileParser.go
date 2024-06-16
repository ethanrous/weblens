package dataProcess

import (
	"errors"
	"fmt"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

// Global exiftool
// var gexift *exiftool.Exiftool
// var gexiftBufferSize int64

type dataProcessController struct {
	media types.MediaRepo
	exif  *exiftool.Exiftool
}

var dpc dataProcessController

// func SetControllers(mediaRepo types.MediaRepo) {
// 	dpc = dataProcessController{
// 		media: mediaRepo,
// 		exif:  media.NewExif(1000 * 1000 * 100),
// 	}
// }

func processMediaFile(t *task) {
	meta := t.metadata.(scanMetadata)
	m := meta.partialMedia
	file := meta.file

	file.AddTask(t)
	defer func(file types.WeblensFile, id types.TaskId) {
		err := file.RemoveTask(id)
		if err != nil {
			util.ShowErr(err)
		}
	}(file, t.TaskId())

	if m == nil {
		t.ErrorAndExit(errors.New("attempted to process nil media"))
		return
	}

	if !file.IsDisplayable(dpc.media) {
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
	if !m.IsCached(file.GetTree()) {
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
	meta := t.metadata.(scanMetadata)
	scanDir := meta.file

	if scanDir.Filename() == ".user_trash" {
		t.taskPool.NotifyTaskComplete(t, t.caster, "No media to scan")
		t.caster.PushTaskUpdate(t.taskId, ScanComplete, types.TaskResult{"execution_time": t.ExeTime()}) // Let any client subscribers know we are done
		t.success("No media to scan")
		return
	}

	// Claim task lock on this file before reading. This
	// prevents lost scans on child files if we were, say,
	// uploading into this directory as a scan comes through.
	// We will block until the upload finishes, then continue this scan
	scanDir.AddTask(t)
	defer func(scanDir types.WeblensFile, id types.TaskId) {
		err := scanDir.RemoveTask(id)
		if err != nil {
			util.ShowErr(err)
		}
	}(scanDir, t.TaskId())

	tp := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)
	util.Info.Printf("Beginning directory scan for %s\n", scanDir.GetAbsPath())

	t.caster.FolderSubToTask(scanDir.ID(), t.TaskId())
	t.caster.FolderSubToTask(scanDir.GetParent().ID(), t.TaskId())

	err := scanDir.LeafMap(func(wf types.WeblensFile) error {
		if wf.IsDir() {
			return nil
			// TODO: Lock directory files while scanning to be able to check what task is using each file
			// wf.AddTask(t)
		}
		// If this file is already being processed, don't queue it again
		fileTask := wf.GetTask()
		if fileTask != nil && fileTask.TaskType() == ScanFileTask {
			return nil
		}

		if !wf.IsDisplayable(dpc.media) {
			return nil
		}

		m := dpc.media.Get(wf.GetContentId())
		if m != nil && m.IsImported() && m.IsCached(wf.GetTree()) {
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
		t.caster.PushTaskUpdate(t.taskId, TaskFailed, types.TaskResult{"failure_note": fmt.Sprintf(
			"%d scans failed", len(errs))}) // Let any client subscribers know we are done
		t.ErrorAndExit(ErrChildTaskFailed)
	}

	t.caster.PushTaskUpdate(t.taskId, ScanComplete, types.TaskResult{"execution_time": t.ExeTime()}) // Let any client subscribers know we are done
	tp.NotifyTaskComplete(t, t.caster)
	t.success()
}

func getScanResult(t *task) types.TaskResult {
	var tp types.TaskPool

	if t.taskPool != nil {
		tp = t.taskPool.GetRootPool()
	}

	result := types.TaskResult{
		"filename": t.metadata.(scanMetadata).file.Filename(),
	}

	if tp != nil {
		complete, total, progress := tp.Status()
		result["percent_progress"] = progress
		result["tasks_complete"] = complete
		result["tasks_total"] = total
		result["task_job_name"] = tp.CreatedInTask().TaskType()
		result["task_job_target"] = tp.CreatedInTask().(*task).metadata.(scanMetadata).file.Filename()
	} else {
		result["task_job_name"] = t.TaskType()
		result["task_job_target"] = t.metadata.(scanMetadata).file.Filename()
	}

	return result
}
