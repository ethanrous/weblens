package jobs

import (
	"errors"
	"path/filepath"
	"slices"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/caster"
	"github.com/ethanrous/weblens/task"
)

func ScanDirectory(t *task.Task) {
	meta := t.GetMeta().(models.ScanMeta)
	meta.Caster = caster.NewSimpleCaster(meta.TaskSubber.(models.ClientManager))

	t.SetErrorCleanup(func(tsk *task.Task) {
		err := t.ReadError()
		meta.Caster.PushTaskUpdate(tsk, models.TaskFailedEvent, task.TaskResult{"error": err.Error()})
	})

	if meta.FileService.IsFileInTrash(meta.File) {
		// Let any client subscribers know we are done
		meta.Caster.PushTaskUpdate(
			t, models.FolderScanCompleteEvent, task.TaskResult{"executionTime": t.ExeTime()},
		)
		t.Success("No media to scan")
		return
	}

	// TODO:
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

	// Create a new task pool for the file scans this directory scan will spawn
	pool := t.GetTaskPool().GetWorkerPool().NewTaskPool(true, t)

	err := meta.FileService.AddTask(meta.File, t)
	if err != nil {
		t.ReqNoErr(err)
	}
	defer func() { err = meta.FileService.RemoveTask(meta.File, t); log.ErrTrace(err) }()

	meta.TaskSubber.FolderSubToTask(meta.File.ID(), t.TaskId())
	meta.TaskSubber.FolderSubToTask(meta.File.GetParentId(), t.TaskId())
	t.SetCleanup(func(tsk *task.Task) {
		// Make sure we finish sending any messages to the client
		// before we close unsubscribe from the task
		meta.Caster.Flush()
		meta.TaskSubber.UnsubTask(tsk.TaskId())
		meta.Caster.Close()
	})

	log.Debug.Printf("Beginning directory scan for %s (%s)", meta.File.GetPortablePath(), meta.File.ID())

	var alreadyFiles []*fileTree.WeblensFileImpl
	var alreadyMedia []*models.Media
	start := time.Now()
	err = meta.File.LeafMap(
		func(wf *fileTree.WeblensFileImpl) error {
			if wf.IsDir() {
				log.Trace.Func(func(l log.Logger) { l.Printf("Skipping file %s, not regular file", wf.AbsPath()) })
				return nil
			}

			if !meta.MediaService.IsFileDisplayable(wf) {
				log.Trace.Func(func(l log.Logger) { l.Printf("Skipping file %s, not displayable", wf.AbsPath()) })
				return nil
			}

			m := meta.MediaService.Get(wf.GetContentId())
			if m != nil && m.IsImported() && meta.MediaService.IsCached(m) {
				if !slices.ContainsFunc(m.FileIds, func(fId fileTree.FileId) bool { return fId == wf.ID() }) {
					err := meta.MediaService.AddFileToMedia(m, wf)
					if err != nil {
						return err
					}
					alreadyFiles = append(alreadyFiles, wf)
					alreadyMedia = append(alreadyMedia, m)
				}
				log.Trace.Func(func(l log.Logger) { l.Printf("Skipping file %s, already imported", wf.AbsPath()) })
				return nil
			}

			subMeta := models.ScanMeta{
				File:         wf,
				FileService:  meta.FileService,
				MediaService: meta.MediaService,
				Caster:       meta.Caster,
			}
			log.Trace.Printf("Dispatching scanFile job for [%s]", wf.AbsPath())
			newT, err := meta.TaskService.DispatchJob(models.ScanFileTask, subMeta, pool)
			if err != nil {
				return err
			}

			newT.SetCleanup(reportSubscanStatus)

			return nil
		},
	)

	log.Debug.Func(func(l log.Logger) {
		l.Printf("Directory scan found files for %s in %s", meta.File.GetPortablePath(), time.Since(start))
	})

	// If the files are already in the media service, we need to update the clients
	// that may be waiting for these files to be processed, but since we won't be
	// adding those to the media service, we need to update the clients now
	meta.Caster.PushFilesUpdate(alreadyFiles, alreadyMedia)

	if err != nil {
		t.ReqNoErr(err)
	}

	pool.SignalAllQueued()

	err = meta.FileService.ResizeDown(meta.File, meta.Caster)
	if err != nil {
		log.ShowErr(err)
	}

	pool.Wait(true, t)

	errs := pool.Errors()
	if len(errs) != 0 {
		// Let any client subscribers know we failed
		meta.Caster.PushTaskUpdate(
			t, models.TaskFailedEvent, task.TaskResult{
				"failedCount": len(errs),
			},
		)
		meta.Caster.PushPoolUpdate(
			pool, models.TaskFailedEvent, task.TaskResult{
				"failedCount": len(errs),
			},
		)
		t.ReqNoErr(werror.WithStack(werror.ErrChildTaskFailed))
	}

	// Let any client subscribers know we are done
	result := getScanResult(t)
	meta.Caster.PushPoolUpdate(pool.GetRootPool(), models.FolderScanCompleteEvent, result)

	log.Debug.Printf("Finished directory scan for %s", meta.File.GetPortablePath())

	t.Success()
}

func ScanFile(t *task.Task) {
	meta := t.GetMeta().(models.ScanMeta)
	// start := time.Now()
	err := ScanFile_(meta, t.ExitIfSignaled)
	// stop := time.Now()
	if err != nil {
		log.Error.Printf("Failed to scan file %s: %s", meta.File.AbsPath(), err)
		t.Fail(err)
	}
	// metrics.MediaProcessTime.Observe(stop.Sub(start).Seconds())

	t.Success()
}

func ScanFile_(meta models.ScanMeta, exitCheck func()) error {
	sw := internal.NewStopwatch("Scan " + meta.File.Filename())
	ext := filepath.Ext(meta.File.Filename())
	if !meta.MediaService.GetMediaTypes().ParseExtension(ext).Displayable {
		log.Error.Printf("Trying to process file with [%s] ext", ext)
		return werror.ErrNonDisplayable
	}
	sw.Lap("Check ext")

	contentId := meta.File.GetContentId()
	if contentId == "" {
		return werror.Errorf("trying to scan file with no content id: %s", meta.File.AbsPath())
	}
	sw.Lap("Check contentId")

	media := models.NewMedia(contentId)
	if slices.ContainsFunc(
		media.GetFiles(), func(fId fileTree.FileId) bool {
			return fId == meta.File.ID()
		},
	) {
		log.Trace.Printf("Media already exists for %s\n", meta.File.Filename())
		return nil
	}
	sw.Lap("New media")

	if meta.PartialMedia == nil {
		meta.PartialMedia = &models.Media{}
	}

	meta.PartialMedia.ContentId = meta.File.GetContentId()
	meta.PartialMedia.FileIds = []fileTree.FileId{meta.File.ID()}
	meta.PartialMedia.Owner = meta.FileService.GetFileOwner(meta.File).GetUsername()

	exitCheck()

	sw.Lap("Checked metadata")
	err := meta.MediaService.LoadMediaFromFile(meta.PartialMedia, meta.File)
	if err != nil {
		return err
	}
	sw.Lap("Loaded media from file")

	exitCheck()

	existingMedia := meta.MediaService.Get(meta.PartialMedia.ID())
	if existingMedia == nil || existingMedia.Height != meta.PartialMedia.Height || existingMedia.
		Width != meta.PartialMedia.Width || len(existingMedia.FileIds) != len(meta.PartialMedia.FileIds) {
		err = meta.MediaService.Add(meta.PartialMedia)
		if err != nil && !errors.Is(err, werror.ErrMediaAlreadyExists) {
			return err
		}
		log.Trace.Printf("Added %s to media service", meta.File.Filename())
		sw.Lap("Added media to service")
	} else {
		log.Debug.Printf("Media already exists for %s\n", meta.File.Filename())
	}

	meta.Caster.PushFileUpdate(meta.File, meta.PartialMedia)
	sw.Lap("Pushed updated file")
	sw.Stop()
	// sw.PrintResults(false)
	log.Trace.Printf("Finished processing %s", meta.File.Filename())

	return nil
}

func reportSubscanStatus(t *task.Task) {
	meta := t.GetMeta().(models.ScanMeta)
	if t.GetTaskPool().IsGlobal() {
		meta.Caster.PushTaskUpdate(t, models.TaskCompleteEvent, getScanResult(t))
	} else {
		meta.Caster.PushPoolUpdate(t.GetTaskPool().GetRootPool(), models.FileScanCompleteEvent, getScanResult(t))
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
			result["taskJobTarget"] = tp.CreatedInTask().GetMeta().(models.ScanMeta).File.Filename()
		} else if tp == nil {
			result["taskJobTarget"] = meta.File.Filename()
		}
	}

	if tp != nil {
		status := tp.Status()
		result["percentProgress"] = status.Progress
		result["tasksComplete"] = status.Complete
		result["tasksFailed"] = status.Failed
		result["tasksTotal"] = status.Total
		result["runtime"] = t.ExeTime()
		if tp.CreatedInTask() != nil {
			result["taskJobName"] = tp.CreatedInTask().JobName()
		}
	} else {
		result["taskJobName"] = t.JobName()
	}

	return result
}
