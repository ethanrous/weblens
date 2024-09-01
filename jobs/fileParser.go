package jobs

import (
	"errors"
	"path/filepath"
	"slices"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/metrics"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/task"
)

func ScanDirectory(t *task.Task) {
	meta := t.GetMeta().(models.ScanMeta)

	if meta.FileService.IsFileInTrash(meta.File) {
		// Let any client subscribers know we are done
		meta.Caster.PushTaskUpdate(
			t, models.ScanCompleteEvent, task.TaskResult{"execution_time": t.ExeTime()},
		)
		t.Success("No media to scan")
		return
	}

	// Claim task lock on this file before reading. This
	// prevents lost scans on child files if we were, say,
	// uploading into this directory as a scan comes through.
	// We will block until the upload finishes, then continue this scan
	// meta.File.AddTask(t)
	// defer func(meta.File *fileTree.WeblensFileImpl, id task.TaskId) {
	// 	err := meta.File.RemoveTask(id)
	// 	if err != nil {
	// 		wlog.ShowErr(err)
	// 	}
	// }(meta.File, t.TaskId())

	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)
	t.SetChildTaskPool(pool)

	err := meta.FileService.AddTask(meta.File, t)
	if err != nil {
		t.ErrorAndExit(err)
	}
	defer func() { err = meta.FileService.RemoveTask(meta.File, t); log.ErrTrace(err) }()

	meta.TaskSubber.FolderSubToPool(meta.File.ID(), pool.GetRootPool().ID())
	meta.TaskSubber.FolderSubToPool(meta.File.GetParentId(), pool.GetRootPool().ID())
	meta.TaskSubber.TaskSubToPool(t.TaskId(), pool.GetRootPool().ID())

	log.Debug.Printf("Beginning directory scan for %s\n", meta.File.GetPortablePath())

	err = meta.File.LeafMap(
		func(wf *fileTree.WeblensFileImpl) error {
			if wf.IsDir() {
				return nil
				// TODO: Lock directory files while scanning to be able to check what task is using each file
				// wf.AddTask(t)
			}

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
				Caster:       meta.Caster,
			}
			newT, err := meta.TaskService.DispatchJob(models.ScanFileTask, subMeta, pool)
			if err != nil {
				return err
			}

			newT.SetCleanup(reportSubscanStatus)

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
		// Let any client subscribers know we failed
		meta.Caster.PushTaskUpdate(
			t, models.TaskFailedEvent, task.TaskResult{
				"failed_count": len(errs),
			},
		)
		meta.Caster.PushPoolUpdate(
			pool, models.TaskFailedEvent, task.TaskResult{
				"failed_count": len(errs),
			},
		)
		t.ErrorAndExit(werror.ErrChildTaskFailed)
	}

	// Let any client subscribers know we are done
	meta.Caster.PushPoolUpdate(
		pool.GetRootPool(), models.ScanCompleteEvent, task.TaskResult{"execution_time": t.ExeTime()},
	)

	result := getScanResult(t)
	meta.Caster.PushTaskUpdate(t, models.SubTaskCompleteEvent, result)
	t.Success()
}

func ScanFile(t *task.Task) {
	meta := t.GetMeta().(models.ScanMeta)
	start := time.Now()
	err := ScanFile_(meta)
	stop := time.Now()
	if err != nil {
		t.ErrorAndExit(err)
	}
	metrics.MediaProcessTime.Observe(stop.Sub(start).Seconds())

	t.Success()
}

func ScanFile_(meta models.ScanMeta) error {
	ext := filepath.Ext(meta.File.Filename())
	if !meta.MediaService.GetMediaTypes().ParseExtension(ext).Displayable {
		log.Error.Printf("Trying to process file with [%s] ext", ext)
		return werror.ErrNonDisplayable
	}

	contentId := models.ContentId(meta.File.GetContentId())
	if contentId == "" {
		return werror.Errorf("trying to scan file with no content id: %s", meta.File.GetAbsPath())
	}

	media := models.NewMedia(contentId)
	if slices.ContainsFunc(
		media.GetFiles(), func(fId fileTree.FileId) bool {
			return fId == meta.File.ID()
		},
	) {
		return nil
	}

	if meta.PartialMedia == nil {
		meta.PartialMedia = &models.Media{}
	}

	meta.PartialMedia.ContentId = models.ContentId(meta.File.GetContentId())
	meta.PartialMedia.FileIds = []fileTree.FileId{meta.File.ID()}
	meta.PartialMedia.Owner = meta.FileService.GetFileOwner(meta.File).GetUsername()

	err := meta.MediaService.LoadMediaFromFile(meta.PartialMedia, meta.File)
	if err != nil {
		return err
	}

	existingMedia := meta.MediaService.Get(meta.PartialMedia.ID())
	if existingMedia == nil || existingMedia.Height != meta.PartialMedia.Height || existingMedia.
		Width != meta.PartialMedia.Width || len(existingMedia.FileIds) != len(meta.PartialMedia.FileIds) {
		err = meta.MediaService.Add(meta.PartialMedia)
		if err != nil && errors.Is(err, werror.ErrMediaAlreadyExists) {
			return err
		}
	}

	meta.Caster.PushFileUpdate(meta.File, meta.PartialMedia)

	return nil
}

func reportSubscanStatus(t *task.Task) {
	meta := t.GetMeta().(models.ScanMeta)
	if t.GetTaskPool().IsGlobal() {
		meta.Caster.PushTaskUpdate(t, models.TaskCompleteEvent, getScanResult(t))
	} else {
		meta.Caster.PushPoolUpdate(t.GetTaskPool().GetRootPool(), models.SubTaskCompleteEvent, getScanResult(t))
	}
}

func getScanResult(t *task.Task) task.TaskResult {
	var tp *task.TaskPool

	if t.GetTaskPool() != nil {
		tp = t.GetTaskPool().GetRootPool()
	}

	var result = task.TaskResult{}
	meta, ok := t.GetMeta().(models.ScanMeta)
	if ok {
		result = task.TaskResult{
			"filename": meta.File.Filename(),
		}
		if tp != nil && tp.CreatedInTask() != nil {
			result["task_job_target"] = tp.CreatedInTask().GetMeta().(models.ScanMeta).File.Filename()
		} else if tp == nil {
			result["task_job_target"] = meta.File.Filename()
		}
	}

	if tp != nil {
		status := tp.Status()
		result["percent_progress"] = status.Progress
		result["tasks_complete"] = status.Complete
		result["tasks_failed"] = status.Failed
		result["tasks_total"] = status.Total
		result["runtime"] = status.Runtime
		if tp.CreatedInTask() != nil {
			result["task_job_name"] = tp.CreatedInTask().JobName()
		}
	} else {
		result["task_job_name"] = t.JobName()
	}

	return result
}
