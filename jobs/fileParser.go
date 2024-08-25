package jobs

import (
	"fmt"

	"github.com/ethrousseau/weblens/comm"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
)

func ScanDirectory(t *task.Task) {
	// if InstanceService.GetLocal().ServerRole() == weblens.BackupServer {
	// 	t.Success()
	// 	return
	// }

	// if types.SERV.MediaRepo == nil {
	// 	t.ErrorAndExit(errors.New("cannot scan directory without valid initilized media repo"))
	// }

	meta := t.GetMeta().(models.ScanMeta)

	if meta.FileService.IsFileInTrash(meta.File) {
		// Let any client subscribers know we are done
		meta.Caster.PushTaskUpdate(
			t, comm.ScanCompleteEvent, task.TaskResult{"execution_time": t.ExeTime()},
		)
		t.Success("No media to scan")
		return
	}

	// Claim task lock on this file before reading. This
	// prevents lost scans on child files if we were, say,
	// uploading into this directory as a scan comes through.
	// We will block until the upload finishes, then continue this scan
	// meta.File.AddTask(t)
	// defer func(meta.File *fileTree.WeblensFile, id task.TaskId) {
	// 	err := meta.File.RemoveTask(id)
	// 	if err != nil {
	// 		wlog.ShowErr(err)
	// 	}
	// }(meta.File, t.TaskId())

	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)
	log.Info.Printf("Beginning directory scan for %s\n", meta.File.GetAbsPath())

	meta.TaskSubber.FolderSubToPool(meta.File.ID(), pool.GetRootPool().ID())
	meta.TaskSubber.FolderSubToPool(meta.File.GetParent().ID(), pool.GetRootPool().ID())

	err := meta.File.LeafMap(
		func(wf *fileTree.WeblensFile) error {
			if wf.IsDir() {
				return nil
				// TODO: Lock directory files while scanning to be able to check what task is using each file
				// wf.AddTask(t)
			}

			// If this file is already being processed, don't queue it again
			// fileTask := wf.GetTask()
			// if fileTask != nil && fileTask.TaskType() == ScanFileTask {
			// 	return nil
			// }

			if !meta.MediaService.IsFileDisplayable(wf) {
				return nil
			}

			m := meta.MediaService.Get(models.ContentId(wf.GetContentId()))
			if m != nil && m.IsImported() && meta.MediaService.IsCached(m) {
				return nil
			}

			subMeta := models.ScanMeta{
				File:         wf,
				FileService:  meta.FileService,
				MediaService: meta.MediaService,
				TaskService:  meta.TaskService,
				Caster:       meta.Caster,
			}
			_, err := meta.TaskService.DispatchJob(models.ScanFileTask, subMeta, pool)
			if err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		t.ErrorAndExit(err)
	}

	pool.SignalAllQueued()

	err = meta.FileService.ResizeDown(meta.File, meta.Caster)
	if err != nil {
		log.ShowErr(err)
	}

	pool.Wait(true)

	errs := pool.Errors()
	if len(errs) != 0 {
		// Let any client subscribers know we are done
		meta.Caster.PushTaskUpdate(
			t, comm.TaskFailedEvent, task.TaskResult{
				"failure_note": fmt.Sprintf(
					"%d scans failed", len(errs),
				),
			},
		)
		t.ErrorAndExit(werror.ErrChildTaskFailed)
	}

	// Let any client subscribers know we are done
	meta.Caster.PushPoolUpdate(
		pool.GetRootPool(), comm.ScanCompleteEvent, task.TaskResult{"execution_time": t.ExeTime()},
	)

	result := getScanResult(t)
	meta.Caster.PushTaskUpdate(t, comm.SubTaskCompleteEvent, result)
	t.Success()
}
