package dataProcess

import (
	"errors"
	"fmt"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func processMediaFile(t *task) {

	meta := t.metadata.(scanMetadata)
	m := meta.partialMedia
	file := meta.file

	// util.Debug.Println("Beginning media file processing for", file.GetAbsPath())

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

	if !file.IsDisplayable() {
		return
	}

	m, err := m.LoadFromFile(file, meta.fileBytes, t)
	if err != nil {
		t.ErrorAndExit(err)
		return
	}

	t.CheckExit()

	err = types.SERV.MediaRepo.Add(m)
	if err != nil {
		t.ErrorAndExit(err)
	}

	t.CheckExit()

	m.Clean()
	if !m.IsCached() {
		t.ErrorAndExit(fmt.Errorf("media scan exiting due to missing cache"))
	}

	if t.caster != nil {
		t.caster.PushFileUpdate(file)
		if t.GetTaskPool().IsGlobal() {
			t.caster.PushTaskUpdate(t, TaskCompleteEvent, getScanResult(t))
		} else {
			t.caster.PushPoolUpdate(t.GetTaskPool().GetRootPool(), SubTaskCompleteEvent, getScanResult(t))
		}
		// t.taskPool.NotifyTaskComplete(t, t.caster)
	} else {
		util.Warning.Println("nil caster in file scan")
	}

	t.success()
}

func scanDirectory(t *task) {
	if types.SERV.InstanceService.GetLocal().ServerRole() == types.Backup {
		t.success()
		return
	}

	if types.SERV.MediaRepo == nil {
		t.ErrorAndExit(types.NewWeblensError("cannot scan directory without valid initilized media repo"))
	}

	meta := t.metadata.(scanMetadata)
	scanDir := meta.file

	if scanDir.Filename() == ".user_trash" {
		t.taskPool.NotifyTaskComplete(t, t.caster, "No media to scan")
		t.caster.PushTaskUpdate(
			t, ScanCompleteEvent, types.TaskResult{"execution_time": t.ExeTime()},
		) // Let any client subscribers know we are done
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

	t.caster.FolderSubToPool(scanDir.ID(), tp.GetRootPool().ID())
	t.caster.FolderSubToPool(scanDir.GetParent().ID(), tp.GetRootPool().ID())

	err := scanDir.LeafMap(
		func(wf types.WeblensFile) error {
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

			if !wf.IsDisplayable() {
				return nil
			}

			m := types.SERV.MediaRepo.Get(wf.GetContentId())
			// if m != nil && m.IsImported() && m.IsCached() {
			//	return nil
			// }
			if m != nil && m.IsImported() {
				return nil
			}

			tp.ScanFile(wf, t.caster)
			return nil
		},
	)

	if err != nil {
		t.ErrorAndExit(err)
	}

	tp.SignalAllQueued()

	err = types.SERV.FileTree.ResizeDown(meta.file, t.caster)
	if err != nil {
		util.ShowErr(err)
	}

	tp.Wait(true)

	errs := tp.Errors()
	if len(errs) != 0 {
		t.caster.PushTaskUpdate(
			t, TaskFailedEvent, types.TaskResult{
				"failure_note": fmt.Sprintf(
					"%d scans failed", len(errs),
				),
			},
		) // Let any client subscribers know we are done
		t.ErrorAndExit(ErrChildTaskFailed)
	}

	// t.caster.PushTaskUpdate(
	// 	t, ScanCompleteEvent, types.TaskResult{"execution_time": t.ExeTime()},
	// )
	t.caster.PushPoolUpdate(
		tp.GetRootPool(), ScanCompleteEvent, types.TaskResult{"execution_time": t.ExeTime()},
	)
	// Let any client subscribers know we are done
	tp.NotifyTaskComplete(t, t.caster)
	t.success()
}

func getScanResult(t *task) types.TaskResult {
	var tp types.TaskPool

	if t.taskPool != nil {
		tp = t.taskPool.GetRootPool()
	}

	var result = types.TaskResult{}
	_, ok := t.metadata.(scanMetadata)
	if ok {
		result = types.TaskResult{
			"filename": t.metadata.(scanMetadata).file.Filename(),
		}
		if tp != nil && tp.CreatedInTask() != nil {
			result["task_job_target"] = tp.CreatedInTask().(*task).metadata.(scanMetadata).file.Filename()
		} else if tp == nil {
			result["task_job_target"] = t.metadata.(scanMetadata).file.Filename()
		}
	}

	if tp != nil {
		status := tp.Status()
		result["percent_progress"] = status.Progress
		result["tasks_complete"] = status.Complete
		result["tasks_total"] = status.Total
		result["runtime"] = status.Runtime
		if tp.CreatedInTask() != nil {
			result["task_job_name"] = tp.CreatedInTask().TaskType()
		}
	} else {
		result["task_job_name"] = t.TaskType()
	}

	return result
}
